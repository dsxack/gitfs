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
var daemonMode = false

var mountCmd = &cobra.Command{
	Use:   "mount <repository> <mountpoint>",
	Short: "Mount git repository into directory",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			repositoryPath = args[0]
			mountPoint     = args[1]
		)

		if daemonMode {
			daemonContext := daemonContextByMountPoint(mountPoint)

			daemonProcess, err := daemonContext.Reborn()
			if err != nil {
				return fmt.Errorf("unable to run daemon process: %w", err)
			}
			if daemonProcess != nil {
				return nil
			}
			defer func() {
				err := daemonContext.Release()
				if err != nil {
					cmd.Printf("unable to release daemon process context: %v", err)
				}
			}()
		}

		repository, err, cleanup := newRepository(cmd, repositoryPath)
		if err != nil {
			return err
		}
		defer cleanup()

		cmd.Println("Mounting filesystem...")
		server, err := fs.Mount(mountPoint, nodes.NewRootNode(repository), &fs.Options{
			MountOptions: fuse.MountOptions{
				Options: mountOptions(repositoryPath, mountPoint),
				FsName:  fmt.Sprintf("gitfs: %s", filepath.Join(repositoryPath, git.GitDirName)),
				Name:    "gitfs",
				Debug:   debug,
			},
		})
		if err != nil {
			return fmt.Errorf("failed to mount filesystem: %w", err)
		}
		cmd.Printf("Filesystem mounted successfully into directory: %s\n", mountPoint)
		go server.Wait()

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt)
		<-sig
		cmd.Println("Received interrupt signal, unmounting...")
		err = server.Unmount()
		if err != nil {
			cmd.Printf("Failed to unmount filesystem: %s\n", err)
		}

		return nil
	},
}

func newRepository(cmd *cobra.Command, repositoryURL string) (*git.Repository, error, func()) {
	dummyCleanup := func() {}

	_, err := os.Stat(repositoryURL)
	if os.IsNotExist(err) {
		clonePath, err := os.MkdirTemp("", "gitfs")
		if err != nil {
			return nil, fmt.Errorf("failed to create temporary directory: %w", err), dummyCleanup
		}
		cmd.Printf("Cloning repository into temporary directory: %s\n", clonePath)
		repository, err := git.PlainClone(clonePath, false, &git.CloneOptions{
			URL: repositoryURL,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to clone repository: %w", err), dummyCleanup
		}
		cmd.Println("Repository cloned successfully")
		return repository, nil, dummyCleanup
	}
	workDirFS := osfs.New(repositoryURL)
	if _, err := workDirFS.Stat(git.GitDirName); err == nil {
		workDirFS, err = workDirFS.Chroot(git.GitDirName)
		if err != nil {
			return nil, fmt.Errorf("failed to chroot to git directory: %w", err), dummyCleanup
		}
	}
	storage := filesystem.NewStorageWithOptions(
		workDirFS,
		cache.NewObjectLRUDefault(),
		filesystem.Options{KeepDescriptors: true},
	)
	cleanup := func() {
		err := storage.Close()
		if err != nil {
			cmd.Println("failed to close storage:", err)
		}
	}
	repository, err := git.Open(storage, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err), cleanup
	}
	return repository, nil, cleanup
}

func init() {
	mountCmd.Flags().BoolVarP(&debug, "verbose", "v", false, "enable verbose output")
	mountCmd.Flags().BoolVarP(&daemonMode, "daemon", "d", false, "run in daemon mode")
}
