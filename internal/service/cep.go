package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/nats-io/nats.go"
	"github.com/spf13/cobra"
)

func Cep(cmd *cobra.Command, _ []string) error {
	configStr := cmd.Flag("config").Value.String()
	cfg, err := serviceCommon(cmd.Context(), configStr)
	if err != nil {
		return err
	}

	return handleCepMsg(cfg)
}

func handleCepMsg(cfg *ConfigService) error {
	_, err := cfg.nc.QueueSubscribe("service.cep", "cep-workers", cepWorkers)
	if err != nil {
		log.Printf("[cep] error subscribing to service.cep: %v", err)
		return err
	}

	log.Println("CEP proxy service listening")
	select {}
}

func cepWorkers(m *nats.Msg) {
	log.Printf("[cep] headers=%v data=%s", m.Header, string(m.Data))

	if _, ok := m.Header["schema"]; ok {
		schema := map[string]string{"cep": "string"}
		b, _ := json.Marshal(schema)
		_ = m.RespondMsg(&nats.Msg{Data: b, Header: nats.Header{"schema": {string(b)}}})
		return
	}

	var req map[string]string
	if err := json.Unmarshal(m.Data, &req); err != nil {
		_ = m.Respond([]byte(`{"error":"bad request"}`))
		return
	}

	cepQ := req["cep"]
	if cepQ == "" {
		_ = m.Respond([]byte(`{"error":"missing cep"}`))
		return
	}

	respMsg, err := queryCEP(cepQ)
	if err != nil {
		log.Printf("[cep] error querying CEP: %v", err)
		_ = m.Respond([]byte(`{"error":"service unavailable"}`))
		return
	}

	_ = m.Respond(respMsg)
}

func queryCEP(cep string) ([]byte, error) {
	url := fmt.Sprintf("https://viacep.com.br/ws/%s/json/", cep)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error en consulta viacep, status: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}
