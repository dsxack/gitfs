package main

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "gitfs",
}

func init() {
	rootCmd.AddCommand(mountCmd)
	rootCmd.AddCommand(umountCmd)
}
