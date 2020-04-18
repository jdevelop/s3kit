package cmd

import "github.com/spf13/cobra"

var lockRoot = &cobra.Command{
	Use:   "lock",
	Short: "Manage object locks",
}

func init() {
	rootCmd.AddCommand(lockRoot)
}
