package cli

import (
	"log"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "aadvcs",
	Short: "aadvcs is a Command Line Interface (CLI) to implement a basic CRDT based Version control system.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Failed to execute command. Reason : %v", err)
	}
}
