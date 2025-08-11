package cmd

import (
	"awesomeProject6/internal/service"

	"github.com/spf13/cobra"
)

// cepCmd represents the cep command
var cepCmd = &cobra.Command{
	Use:   "cep",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: service.Cep,
}

func init() {
	rootCmd.AddCommand(cepCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// cepCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// cepCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
