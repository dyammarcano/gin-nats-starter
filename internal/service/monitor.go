package service

import "github.com/spf13/cobra"

func Monitor(cmd *cobra.Command, _ []string) error {
	configStr := cmd.Flag("config").Value.String()
	_, err := serviceCommon(cmd.Context(), configStr)
	if err != nil {
		return err
	}

	return nil
}
