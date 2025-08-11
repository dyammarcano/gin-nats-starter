package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
	"github.com/spf13/cobra"
)

func Api(cmd *cobra.Command, _ []string) error {
	configStr := cmd.Flag("config").Value.String()
	cfg, err := serviceCommon(cmd.Context(), configStr)
	if err != nil {
		return err
	}

	r := gin.Default()

	proxy := func(subject string, timeout time.Duration) gin.HandlerFunc {
		return func(c *gin.Context) {
			var body map[string]any
			if err := c.BindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			b, _ := json.Marshal(body)
			msg := &nats.Msg{Subject: subject, Data: b, Header: nats.Header{}}
			if s := c.GetHeader("schema"); s != "" {
				msg.Header.Set("schema", "1")
			}

			if d := c.GetHeader("data"); d != "" {
				msg.Header.Set("data", d)
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
			defer cancel()

			resp, err := cfg.nc.RequestMsgWithContext(ctx, msg)
			if err != nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
				return
			}

			c.Data(http.StatusOK, "application/json", resp.Data)
		}
	}

	r.POST("/lookup/cep", proxy("service.cep", 2*time.Second))
	r.POST("/lookup/clima", proxy("service.clima", 2*time.Second))
	r.POST("/validate/identity", proxy("service.identity", 3*time.Second))

	port := cfg.Port
	log.Printf("starting api on :%d", port)
	return r.Run(fmt.Sprintf(":%d", port))
}
