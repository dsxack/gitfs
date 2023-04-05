package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"path/filepath"
	"syscall"
)

var umountCmd = &cobra.Command{
	Use:   "umount <mountPoint>",
	Short: "Unmount git repository from directory",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mountPoint, err := filepath.Abs(args[0])
		if err != nil {
			return fmt.Errorf("unable to get absolute path for '%s': %w", args[0], err)
		}

		daemonContext, err := daemonContextByMountPoint(mountPoint)
		if err != nil {
			return err
		}

		daemonProcess, err := daemonContext.Search()
		if err != nil {
			return fmt.Errorf("unable to umount '%s': maybe it is not mounted", mountPoint)
		}

		err = daemonProcess.Signal(syscall.SIGINT)
		if err != nil {
			return fmt.Errorf("unable to interrupt daemon process: %w", err)
		}

		return nil
	},
}
