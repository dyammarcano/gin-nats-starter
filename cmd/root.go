package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "github.com/dyammarcano/gin-nats-starter",
	Short: "A brief description of your application",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().String("config", "", "config file (default is $HOME/.github.com/dyammarcano/gin-nats-starter.yaml)")
}
