package cmd

import (
	"awesomeProject6/internal/service"

	"github.com/spf13/cobra"
)

// cnpjCmd represents the cnpj command
var cnpjCmd = &cobra.Command{
	Use:   "cnpj",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: service.Cnpj,
}

func init() {
	rootCmd.AddCommand(cnpjCmd)
}
