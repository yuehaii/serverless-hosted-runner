// Package app of micro service entries
package app

import (
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:          "serverless-runner",
		Short:        "Serverless hosted runner.",
		Long:         "Serverless hosted runner running in serverless cloud plarform.",
		SilenceUsage: true,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
	}
	rootCmd.AddCommand(dispacherCmd)
	rootCmd.AddCommand(runnerCmd)
	return rootCmd
}
