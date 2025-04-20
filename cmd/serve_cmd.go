package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var serveCmd = cobra.Command{
	Use:  "serve",
	Long: "Start the server to handle incoming requests",
	Run: func(cmd *cobra.Command, args []string) {
		serve(cmd.Context())
	},
}

func serve(_ context.Context) {
	fmt.Println("Server started")
}
