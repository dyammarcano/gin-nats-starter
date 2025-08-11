package service

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
	"github.com/nats-io/nats.go"
	"github.com/spf13/cobra"
)

type Route struct {
	id     string
	target string
	method string
}

type Proxy struct {
	port             string
	engine           *gin.Engine
	registeredRoutes map[string]*Route
	nc               *nats.Conn
}

func Api(cmd *cobra.Command, _ []string) error {
	configStr := cmd.Flag("config").Value.String()
	cfg, err := serviceCommon(cmd.Context(), configStr)
	if err != nil {
		return err
	}

	doc, err := loadOpenAPI(cfg.OpenApiPath)
	if err != nil {
		return err
	}

	px := &Proxy{
		registeredRoutes: make(map[string]*Route),
		engine:           setupRouter(),
		port:             fmt.Sprintf(":%d", cfg.Port),
		nc:               cfg.nc,
	}

	for p, pathItem := range doc.Spec().Paths.Paths {
		if err := checkPathItem(p, pathItem); err != nil {
			return err
		}

		px.registerRoute(p, pathItem.Get, "GET")
		px.registerRoute(p, pathItem.Post, "POST")
		px.registerRoute(p, pathItem.Put, "PUT")
		px.registerRoute(p, pathItem.Delete, "DELETE")
		px.registerRoute(p, pathItem.Options, "OPTIONS")
		px.registerRoute(p, pathItem.Head, "HEAD")
		px.registerRoute(p, pathItem.Patch, "PATCH")
	}

	port := cfg.Port
	log.Printf("starting api on :%d", port)
	return px.engine.Run(fmt.Sprintf(":%d", port))
}

func setupRouter() *gin.Engine {
	r := gin.New()
	setupMiddleware(r)
	setupHealthEndpoints(r)
	return r
}

func setupMiddleware(r *gin.Engine) {
	r.Use(gin.Recovery())

	methods := strings.Split(os.Getenv("CORS_ALLOW_METHODS"), ",")
	if len(methods) == 0 || methods[0] == "" {
		methods = []string{"GET", "POST"}
	}

	headers := strings.Split(os.Getenv("CORS_ALLOW_HEADERS"), ",")
	if len(headers) == 0 || headers[0] == "" {
		headers = []string{"Content-Type", "Authorization"}
	}

	origins := strings.Split(os.Getenv("CORS_ALLOW_ORIGINS"), ",")
	if len(origins) == 0 || origins[0] == "" {
		origins = []string{"https://example.com"}
	}

	r.Use(cors.New(cors.Config{
		AllowMethods: methods,
		AllowHeaders: headers,
		AllowOrigins: origins,
	}))
}

func setupHealthEndpoints(r *gin.Engine) {
	r.GET("/heartbeat", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"message":   "Service is running",
			"timestamp": time.Now().Unix(),
		})
	})
}

func (p *Proxy) registerRoute(pathRoute string, operation *spec.Operation, method string) {
	if operation != nil {
		key := fmt.Sprintf("%s-%s", method, pathRoute)
		if p.registeredRoutes[key] != nil {
			return
		}

		s := sha256.New()
		s.Write([]byte(key))

		subject := getExtensionString(operation.Extensions["x-nats-subject"])
		timeout := getExtensionDuration(operation.Extensions["x-timeout"], 2*time.Second)

		p.registeredRoutes[key] = &Route{
			id:     fmt.Sprintf("%x-%x", s.Sum(nil)[0:3], s.Sum(nil)[5:7]),
			target: pathRoute,
			method: method,
		}

		handler := proxyNats(context.Background(), p.nc, subject, timeout)
		p.engine.Handle(method, pathRoute, handler)
	}
}

func checkPathItem(path string, pathItem spec.PathItem) error {
	if pathItem.Get == nil && pathItem.Post == nil && pathItem.Put == nil &&
		pathItem.Delete == nil && pathItem.Options == nil && pathItem.Head == nil && pathItem.Patch == nil {
		return fmt.Errorf("path item is empty for path: %s", path)
	}

	return nil
}

func loadOpenAPI(filePath string) (*loads.Document, error) {
	doc, err := loads.Spec(filePath)
	if err != nil {
		return nil, fmt.Errorf("error loading Swagger file: %v", err)
	}

	if doc.Spec().Paths == nil {
		return nil, fmt.Errorf("swagger spec is missing paths")
	}

	if doc.Spec().Info == nil {
		return nil, fmt.Errorf("swagger spec is missing info")
	}

	if doc.Spec().Info.Title == "" {
		return nil, fmt.Errorf("swagger spec is missing info title")
	}

	return doc, nil
}

func proxyNats(ctx context.Context, nc *nats.Conn, subject string, timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctxTimeout, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		msg := &nats.Msg{Subject: subject, Data: []byte("empty"), Header: nats.Header{}}
		if s := c.GetHeader("schema"); s != "" {
			msg.Header.Set("schema", "1")
			resp, err := nc.RequestMsgWithContext(ctxTimeout, msg)
			if err != nil {
				if errors.Is(err, nats.ErrNoResponders) {
					c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service unavailable to process request"})
					return
				}
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
				return
			}
			c.Data(http.StatusOK, "application/json", resp.Data)
			return
		}

		var body map[string]any
		if err := c.BindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		b, _ := json.Marshal(body)
		msg.Data = b

		if d := c.GetHeader("data"); d != "" {
			msg.Header.Set("data", d)
		}

		resp, err := nc.RequestMsgWithContext(ctx, msg)
		if err != nil {
			if errors.Is(err, nats.ErrNoResponders) {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service unavailable to process request"})
				return
			}
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		c.Data(http.StatusOK, "application/json", resp.Data)
	}
}

func getExtensionString(ext interface{}) string {
	if ext == nil {
		return ""
	}
	if s, ok := ext.(string); ok {
		return s
	}
	return ""
}

func getExtensionDuration(ext interface{}, def time.Duration) time.Duration {
	if ext == nil {
		return def
	}
	if s, ok := ext.(string); ok {
		if d, err := time.ParseDuration(s); err == nil {
			return d
		}
	}
	return def
}
