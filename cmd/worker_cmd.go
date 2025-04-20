package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var workerCmd = cobra.Command{
	Use:  "worker",
	Long: "Start the worker to process tasks",
	Run: func(cmd *cobra.Command, args []string) {
		worker(cmd.Context())
	},
}

func worker(_ context.Context) {
	fmt.Println("Worker started")
}
