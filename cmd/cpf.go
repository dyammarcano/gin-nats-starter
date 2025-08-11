package cmd

import (
	"awesomeProject6/internal/service"

	"github.com/spf13/cobra"
)

// cpfCmd represents the cpf command
var cpfCmd = &cobra.Command{
	Use:   "cpf",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: service.Cpf,
}

func init() {
	rootCmd.AddCommand(cpfCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// cpfCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// cpfCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
