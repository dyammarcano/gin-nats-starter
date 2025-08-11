package service

import (
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

type Identity struct {
	gorm.Model
	UUID     string `gorm:"uniqueIndex"`
	CPF      string `gorm:"index"`
	CNPJ     string `gorm:"index"`
	Name     string
	Verified bool
}

func Cnpj(cmd *cobra.Command, args []string) error {
	configStr := cmd.Flag("config").Value.String()
	cfg, err := serviceCommon(configStr)
	if err != nil {
		return err
	}

	return nil
}
