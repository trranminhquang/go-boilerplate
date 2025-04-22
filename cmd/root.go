package cmd

import (
	"sync"

	"github.com/spf13/cobra"
)

var (
	configFile = ""
	watchDir   = ""
	runAll     = false
)

var rootCmd = cobra.Command{
	Use:   "app",
	Short: "Go Boilerplate Application",
	Long:  "A simple boilerplate for building REST APIs and workers in Go",
	Run: func(cmd *cobra.Command, args []string) {
		if runAll {
			var wg sync.WaitGroup
			wg.Add(2)

			// Start worker in a separate goroutine
			go func() {
				defer wg.Done()
				startWorker(cmd.Context())
			}()

			// Start server in a separate goroutine
			go func() {
				defer wg.Done()
				startServer(cmd.Context())
			}()

			// Wait for both to complete (which won't happen unless context is canceled)
			wg.Wait()
		} else {
			// By default, only run the server
			startServer(cmd.Context())
		}
	},
}

// RootCommand returns the root command for the application
func RootCommand() *cobra.Command {
	rootCmd.AddCommand(&serveCmd, &workerCmd)
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "base configuration file to load")
	rootCmd.PersistentFlags().StringVarP(&watchDir, "config-dir", "d", "", "directory containing a sorted list of config files to watch for changes")
	rootCmd.Flags().BoolVar(&runAll, "all", false, "run both server and worker")

	return &rootCmd
}
