package cmd

import (
	"os"
	"path"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "verman",
	Short: "A plugin for Spin that makes it easy to manage different versions of the Spin CLI.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(setCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(lsCmd)
	rmCmd.AddCommand(rmAllCmd)
	rmCmd.AddCommand(rmCurrentCmd)
	rootCmd.AddCommand(rmCmd)
}

func exists(path string) (bool, error) {
	// If the path does exist...
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	// If the path doesn't exist...
	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

func getVermanDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return path.Join(homeDir, ".spin_verman"), nil
}
