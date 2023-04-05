package main

import (
	"crypto/md5"
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/sevlyar/go-daemon"
	"log"
	"os"
	"path/filepath"
)

var (
	homeDir string
	logsDir string
	pidsDir string
)

func init() {
	userHomeDir, err := homedir.Dir()
	if err != nil {
		log.Fatalf("failed to get user home path: %v", err)
	}

	homeDir = userHomeDir + "/.gitfs"
	logsDir = homeDir + "/logs"
	pidsDir = homeDir + "/pids"

	for _, dir := range []string{
		homeDir,
		logsDir,
		pidsDir,
	} {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			log.Fatalf("failed to create home path: %v", err)
		}
	}
}

func daemonContextByMountPoint(mountPoint string) (daemon.Context, error) {
	mountPoint, err := filepath.Abs(mountPoint)
	if err != nil {
		return daemon.Context{}, fmt.Errorf("unable to get absolute path for '%s': %w", mountPoint, err)
	}

	return daemon.Context{
		PidFileName: pidFileName(mountPoint),
		PidFilePerm: 0644,
		LogFileName: logFileName(mountPoint),
		LogFilePerm: 0640,
		WorkDir:     "./",
		Args:        os.Args,
	}, nil
}

func mountPointHash(mountPoint string) string {
	hash := md5.Sum([]byte(mountPoint))
	return fmt.Sprintf("%x", hash)
}

func pidFileName(mountPoint string) string {
	return fmt.Sprintf("%s/%s-%s.pid", pidsDir, filepath.Base(mountPoint), mountPointHash(mountPoint))
}

func logFileName(mountPoint string) string {
	return fmt.Sprintf("%s/%s-%s.log", logsDir, filepath.Base(mountPoint), mountPointHash(mountPoint))
}
