package main

import (
	"fmt"
	"github.com/dsxack/gitfs/nodes"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"path/filepath"
)

var debug bool

var mountCmd = &cobra.Command{
	Use:   "mount <repository> <mountpoint>",
	Short: "Mount git repository into directory",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			repositoryPath = args[0]
			mountPoint     = args[1]
		)

		workDirFS := osfs.New(repositoryPath)
		if _, err := workDirFS.Stat(git.GitDirName); err == nil {
			workDirFS, err = workDirFS.Chroot(git.GitDirName)
			if err != nil {
				return err
			}
		}
		storage := filesystem.NewStorageWithOptions(
			workDirFS,
			cache.NewObjectLRUDefault(),
			filesystem.Options{KeepDescriptors: true},
		)
		repository, err := git.Open(storage, nil)
		if err != nil {
			return err
		}
		defer storage.Close()

		volumeName := fmt.Sprintf(
			"%s (%s)",
			filepath.Base(mountPoint),
			filepath.Base(repositoryPath),
		)

		server, err := fs.Mount(mountPoint, nodes.NewRootNode(repository), &fs.Options{
			MountOptions: fuse.MountOptions{
				Options: []string{
					"volname=" + volumeName,
				},
				FsName: fmt.Sprintf("gitfs: %s", filepath.Join(repositoryPath, git.GitDirName)),
				Name:   "gitfs",
				Debug:  debug,
			},
		})
		if err != nil {
			return err
		}

		go server.Wait()

		sig := make(chan os.Signal)
		signal.Notify(sig, os.Interrupt)
		select {
		case <-sig:
			server.Unmount()
		}

		return nil
	},
}

func init() {
	mountCmd.Flags().BoolVarP(&debug, "debug", "d", false, "enable debug output")
}
