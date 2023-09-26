package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"runtime/debug"
	"time"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(getVersionString())
	},
}

func getVersionString() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "gitfs: unknown version"
	}
	var revision string
	var lastCommit time.Time

	for _, kv := range info.Settings {
		switch kv.Key {
		case "vcs.revision":
			revision = kv.Value
		case "vcs.time":
			lastCommit, _ = time.Parse(time.RFC3339, kv.Value)
		}
	}
	if revision == "" {
		return fmt.Sprintf("gitfs: version %s", info.Main.Version)
	}
	return fmt.Sprintf(
		"gitfs: version %s, build %s",
		revision,
		lastCommit,
	)
}
