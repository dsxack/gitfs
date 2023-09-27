package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"runtime/debug"
)

var version string

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println(getVersionString())
	},
}

func getVersionString() string {
	if version != "" {
		return fmt.Sprintf("gitfs: v%s", version)
	}

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "gitfs: unknown version"
	}

	return fmt.Sprintf("gitfs: %s", info.Main.Version)
}
