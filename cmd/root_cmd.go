package cmd

import "github.com/spf13/cobra"

var (
	configFile = ""
	watchDir   = ""
)

var rootCmd = cobra.Command{
	Run: func(cmd *cobra.Command, args []string) {
		worker(cmd.Context())
		serve(cmd.Context())
	},
}

func RootCommand() *cobra.Command {
	rootCmd.AddCommand(&serveCmd, &workerCmd)
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "base configuration file to load")
	rootCmd.PersistentFlags().StringVarP(&watchDir, "config-dir", "d", "", "directory containing a sorted list of config files to watch for changes")

	return &rootCmd
}
