package main

import (
	"fmt"
	"github.com/dsxack/gitfs/nodes"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/spf13/cobra"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"syscall"
)

var daemonModeFlag = false
var verboseLevel int

func init() {
	mountCmd.Flags().CountVarP(&verboseLevel, "verbose", "v", "enable verbose output")
	mountCmd.Flags().BoolVarP(&daemonModeFlag, "daemon", "d", false, "run in daemon mode")
}

var mountCmd = &cobra.Command{
	Use:   "mount <repository> <mountpoint>",
	Short: "Mount git repository into directory",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			repositoryPath = args[0]
			mountPoint     = args[1]
		)
		setupLogger()

		if daemonModeFlag {
			daemonContext, err := daemonContextByMountPoint(mountPoint)
			if err != nil {
				return err
			}

			daemonProcess, err := daemonContext.Reborn()
			if err != nil {
				return fmt.Errorf("unable to run daemon process: %w", err)
			}
			if daemonProcess != nil {
				cmd.Printf("Running in daemon mode, logs could be discovered in %s\n", daemonContext.LogFileName)
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
				Debug:   verboseLevel > 2,
			},
		})
		if err != nil {
			return fmt.Errorf("failed to mount filesystem: %w", err)
		}
		cmd.Printf("Filesystem successfully mounted into directory: %s\n", mountPoint)

		go server.Wait()

		sigC := make(chan os.Signal, 1)
		signal.Notify(
			sigC,
			syscall.SIGTERM,
			syscall.SIGINT,
			syscall.SIGQUIT,
		)

		for {
			sig := <-sigC

			switch sig {
			case syscall.SIGTERM, syscall.SIGINT:
				cmd.Println("Received " + sig.String() + ", unmounting...")
				err := server.Unmount()
				if err != nil {
					cmd.Printf("Failed to unmount filesystem: %s\n", err)
					continue
				}
				return nil

			case syscall.SIGQUIT:
				cmd.Printf("Go version: %s %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
				_ = pprof.Lookup("goroutine").WriteTo(os.Stdout, 1)
				continue
			}
		}
	},
}

func newRepository(cmd *cobra.Command, repositoryURL string) (*git.Repository, error, func()) {
	dummyCleanup := func() {}

	_, err := os.Stat(repositoryURL)
	if os.IsNotExist(err) {
		cmd.Printf("Cloning repository %s into memory\n", repositoryURL)
		storage := memory.NewStorage()
		r, err := git.Clone(storage, nil, &git.CloneOptions{
			NoCheckout: true,
			URL:        repositoryURL,
			Progress:   cmd.OutOrStderr(),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to clone repository: %w", err), dummyCleanup
		}
		return r, nil, dummyCleanup
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

func mountOptions(repositoryPath, mountPoint string) []string {
	var options []string

	//goland:noinspection GoBoolExpressions
	if runtime.GOOS == "darwin" {
		volumeName := fmt.Sprintf(
			"%s (%s)",
			filepath.Base(mountPoint),
			filepath.Base(repositoryPath),
		)
		options = append(options, "volname="+volumeName)
	}

	return options
}

func setupLogger() {
	logLevel := slog.LevelError
	if verboseLevel > 0 {
		logLevel = slog.LevelInfo
	}
	if verboseLevel > 1 {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(
		os.Stderr,
		&slog.HandlerOptions{
			Level: logLevel,
		},
	))
	slog.SetDefault(logger)
}
