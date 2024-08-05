package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
)

// Because the subcommands for 'rm' utilize some of the same methods, it didn't seem to make sense to separate them into a subcommand directory

var rmCmd = &cobra.Command{
	Use:   "rm [version|*]",
	Short: "Removes the specified Spin version from the local directory.",
	Long:  "Removes the specified Spin version from the local directory. Only removes the relevant Spin binary located in the \"~/.spin_verman/versions\" directory.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("you must indicate which version of Spin you wish to delete or use '*' to delete all versions")
		}

		version := args[0]

		if version == "all" {
			return rmAllCmd.RunE(cmd, args)
		}

		if !strings.HasPrefix(version, "v") {
			version = "v" + version
		}

		if err := rm(version); err != nil {
			return err
		}

		return nil
	},
}

var rmCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Removes the alternate Spin version, reverting back to the root version of Spin.",
	Long:  "Removes the alternate Spin version, reverting back to the root version of Spin, but preserving all other versions of Spin downloaded locally.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := rm("current_version"); err != nil {
			return err
		}

		return nil
	},
}

var rmAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Removes all Spin versions from the local directory.",
	Long:  "Removes all Spin versions from the local directory. Only removes the Spin binaries located in the \"~/.spin_verman/versions\" directory.",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Print("Are you sure you want to delete all Spin versions?\nType \"y\", \"yes\", or any other key to cancel: ")
		input := bufio.NewScanner(os.Stdin)
		input.Scan()
		output := strings.ToLower(input.Text())

		if output == "y" || output == "yes" {
			if err := rmAll(); err != nil {
				return err
			}
		}

		return nil
	},
}

func rm(version string) error {
	vermanDir, err := getVermanDir()
	if err != nil {
		return err
	}

	filePath := path.Join(vermanDir, "versions", version)

	if err := os.RemoveAll(filePath); err != nil {
		return err
	}

	return nil
}

func rmAll() error {
	versionString, err := list()
	if err != nil {
		return err
	}

	for _, version := range strings.Split(versionString, "\n") {
		if err := rm(version); err != nil {
			return err
		}
	}

	// The list method doesn't return the "current_version" directory, so we need to manually delete it
	if err := rm("current_version"); err != nil {
		return err
	}

	return nil
}
