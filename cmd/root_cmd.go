package cmd

import "github.com/spf13/cobra"

var rootCmd = cobra.Command{
	Run: func(cmd *cobra.Command, args []string) {
		serve(cmd.Context())
		worker(cmd.Context())
	},
}

func RootCommand() *cobra.Command {
	rootCmd.AddCommand(&serveCmd, &workerCmd)

	return &rootCmd
}
