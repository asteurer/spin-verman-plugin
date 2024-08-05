package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/spf13/cobra"
)

var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Sets Spin to the requested version.",
	Long:  "Sets Spin to the requested version, and will download the binary for the requested version if not found locally.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("you must indicate the version of Spin you wish to set")
		}

		version := args[0]

		if !strings.HasPrefix(version, "v") {
			version = "v" + version
		}

		vermanDir, err := getVermanDir()
		if err != nil {
			return err
		}

		if err := downloadSpin(vermanDir, version); err != nil {
			return err
		}

		if err = updateSpinBinary(vermanDir, version); err != nil {
			return err
		}

		fmt.Printf("Spin has been updated to version %s\n", version)
		return nil
	},
}

func updateSpinBinary(directory, version string) error {
	directory = path.Join(directory, "versions")

	if err := os.MkdirAll(path.Join(directory, "current_version"), 0755); err != nil {
		return err
	}

	symLinkDir := path.Join(directory, "current_version")

	// Removing old SymLink, returning an error only if the error is not a 'file does not exist' error
	if err := os.Remove(path.Join(symLinkDir, "spin")); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove old symlink: %v", err)
		}
	}

	if err := os.Symlink(path.Join(directory, version, "spin"), path.Join(symLinkDir, "spin")); err != nil {
		return err
	}

	testSpinVersionCmd := exec.Command("spin", "--version")
	outputBytes, err := testSpinVersionCmd.CombinedOutput()
	if err != nil {
		return err
	}

	if !strings.Contains(string(outputBytes), version[1:]) {
		return fmt.Errorf("it looks like the version of the current Spin executable does not match what was requested, so please check to make sure the path %q is prepended to your path", symLinkDir)
	}

	return nil
}
