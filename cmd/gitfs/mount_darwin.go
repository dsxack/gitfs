package main

import (
	"fmt"
	"path/filepath"
)

func mountOptions(repositoryPath, mountPoint string) []string {
	volumeName := fmt.Sprintf(
		"%s (%s)",
		filepath.Base(mountPoint),
		filepath.Base(repositoryPath),
	)
	return []string{
		"volname=" + volumeName,
	}
}
