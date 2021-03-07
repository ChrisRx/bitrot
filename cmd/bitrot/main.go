package main

import (
	"log"

	"github.com/spf13/cobra"
)

var (
	dbPath string

	rootCmd = &cobra.Command{
		Use: "bitrot",
	}
)

func init() {
	rootCmd.PersistentFlags().StringVar(&dbPath, "dir", "data", "db path")
}

func main() {
	rootCmd.AddCommand(
		corruptCommand,
		fixCommand,
		scanCommand,
	)
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
