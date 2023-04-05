package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"syscall"
)

var umountCmd = &cobra.Command{
	Use:   "umount <mountPoint>",
	Short: "Unmount git repository from directory",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			mountPoint = args[0]
		)

		daemonContext := daemonContextByMountPoint(mountPoint)

		daemonProcess, err := daemonContext.Search()
		if err != nil {
			return fmt.Errorf("unable to search daemon process: %w", err)
		}

		err = daemonProcess.Signal(syscall.SIGINT)
		if err != nil {
			return fmt.Errorf("unable to interrupt daemon process: %w", err)
		}

		state, err := daemonProcess.Wait()
		if err != nil {
			return fmt.Errorf("unable to wait for daemon process: %w", err)
		}

		if !state.Success() {
			return fmt.Errorf("daemon process exited with non-zero exit code: %d", state.ExitCode())
		}

		cmd.Println("Successfully unmounted filesystem.")

		return nil
	},
}