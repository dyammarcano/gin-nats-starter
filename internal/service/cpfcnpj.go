package service

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/spf13/cobra"
)

var notAccepted []string

type (
	CPF struct {
		cpfNumber []int
	}

	CNPJ struct {
		cnpjNumber []int
	}

	Response struct {
		CPF     string
		IsValid bool
		Origin  string
	}
)

func init() {
	notAccepted = make([]string, 0)
	for i := 0; i < 10; i++ {
		value := strings.Repeat(strconv.Itoa(i), 11)
		notAccepted = append(notAccepted, value)
	}
}

func CpfCnpj(cmd *cobra.Command, _ []string) error {
	configStr := cmd.Flag("config").Value.String()
	cfg, err := serviceCommon(cmd.Context(), configStr)
	if err != nil {
		return err
	}

	_, err = cfg.nc.QueueSubscribe("service.cpfcnpj", "cpfcnpj-workers", cpfcnpjWorkers)
	if err != nil {
		log.Printf("[cpfcnpj] error subscribing to service.cpfcnpj: %v", err)
		return err
	}

	log.Println("CPFCNPJ proxy service listening")
	select {}
}

func cpfcnpjWorkers(m *nats.Msg) {
	log.Printf("[cpfcnpj] headers=%v data=%s", m.Header, string(m.Data))

	obj := &CPF{}

	if _, ok := m.Header["schema"]; ok {
		schema := map[string]string{"cpfcnpj": "string"}
		b, _ := json.Marshal(schema)
		_ = m.RespondMsg(&nats.Msg{Data: b, Header: nats.Header{"schema": {string(b)}}})
		return
	}

	var req map[string]string
	if err := json.Unmarshal(m.Data, &req); err != nil {
		_ = m.Respond([]byte(`{"error":"bad request"}`))
		return
	}

	cepQ := req["cpfcnpj"]
	if cepQ == "" {
		_ = m.Respond([]byte(`{"error":"missing cpfcnpj"}`))
		return
	}

	if !obj.Validate(cepQ) {
		_ = m.Respond([]byte(`{"error":"document invalid"}`))
	}

	data, _ := json.Marshal(struct {
		Document string `json:"document"`
		IsValid  bool   `json:"is_valid"`
		Origin   string `json:"origin"`
	}{
		Document: obj.Format(cepQ),
		IsValid:  obj.Validate(cepQ),
		Origin:   obj.CheckOrigin(cepQ),
	})

	_ = m.Respond(data)
}

func (c *CPF) Generate() string {
	rand.Seed(time.Now().UnixNano())
	number := make([]int, 9)
	for i := 0; i < 9; i++ {
		number[i] = rand.Intn(10)
	}
	number = append(number, c.calculateFirstDigit(number))
	number = append(number, c.calculateSecondDigit(number))
	return c.maskCPF(number)
}

func (c *CPF) maskCPF(values []int) string {
	cpf := ""
	for _, item := range values {
		cpf += strconv.Itoa(item)
	}
	cpf = strings.ReplaceAll(cpf, "-", "")
	return fmt.Sprintf("%s.%s.%s-%s", cpf[:3], cpf[3:6], cpf[6:9], cpf[9:])
}

func (c *CPF) clean(values string) {
	c.cpfNumber = nil
	values = strings.ReplaceAll(values, "-", "")
	for _, item := range values {
		digit, err := strconv.Atoi(string(item))
		if err == nil {
			c.cpfNumber = append(c.cpfNumber, digit)
		}
	}
}

func (c *CPF) calculateFirstDigit(values []int) int {
	sum := 0
	for i := 0; i < 9; i++ {
		sum += values[i] * (10 - i)
	}
	rest := (sum * 10) % 11
	if rest == 10 || rest == 11 {
		rest = 0
	}
	return rest
}

func (c *CPF) validate(values []int) bool {
	return c.calculateFirstDigit(values) == values[9] && c.calculateSecondDigit(values) == values[10]
}

func (c *CPF) calculateSecondDigit(values []int) int {
	sum := 0
	for i := 0; i < 10; i++ {
		sum += values[i] * (11 - i)
	}
	rest := (sum * 10) % 11
	if rest == 10 || rest == 11 {
		rest = 0
	}
	return rest
}

func (c *CPF) Validate(values string) bool {
	c.clean(values)
	return c.isAccepted(values) && c.length(c.cpfNumber) && c.validate(c.cpfNumber)
}

func (c *CPF) isAccepted(values string) bool {
	cpf := strings.ReplaceAll(values, "-", "")
	for _, notAccepted := range notAccepted {
		if cpf == notAccepted {
			return false
		}
	}
	return true
}

func (c *CPF) length(values []int) bool {
	return len(values) == 11
}

func (c *CPF) CheckOrigin(values string) string {
	c.clean(values)
	switch c.cpfNumber[8] {
	case 0:
		return "Rio Grande do Sul"
	case 1:
		return "Distrito Federal, Goiás, Mato Grosso do Sul e Tocantins"
	case 2:
		return "Pará, Amazonas, Acre, Amapá, Rondônia e Roraima"
	case 3:
		return "Ceará, Maranhão e Piauí"
	case 4:
		return "Pernambuco, Rio Grande do Norte, Paraíba e Alagoas"
	case 5:
		return "Bahia e Sergipe"
	case 6:
		return "Minas Gerais"
	case 7:
		return "Rio de Janeiro e Espírito Santo"
	case 8:
		return "São Paulo"
	case 9:
		return "Paraná e Santa Catarina"
	default:
		return ""
	}
}

func (c *CPF) Format(s string) string {
	c.clean(s)
	return c.maskCPF(c.cpfNumber)
}
