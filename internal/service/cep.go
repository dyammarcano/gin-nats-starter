package service

import (
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

type CEP struct {
	gorm.Model
	CEP    string `gorm:"uniqueIndex"`
	Street string
	City   string
	State  string
}

func Cep(cmd *cobra.Command, args []string) error {
	configStr := cmd.Flag("config").Value.String()
	cfg, err := serviceCommon(configStr)
	if err != nil {
		return err
	}

	return nil
}
