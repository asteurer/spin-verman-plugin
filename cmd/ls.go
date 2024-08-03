package cmd

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "Lists all Spin versions downloaded locally.",
	RunE: func(cmd *cobra.Command, args []string) error {
		output, err := list()
		if err != nil {
			return err
		}

		fmt.Println(output)

		return nil
	},
}

func list() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	versionPath := path.Join(homeDir, ".spin_verman", "versions")

	pathExists, err := exists(versionPath)
	if err != nil {
		return "", err
	}

	if !pathExists {
		return "", nil
	}

	files, err := os.ReadDir(versionPath)
	if err != nil {
		return "", err
	}

	var output []string

	for _, file := range files {
		if strings.HasPrefix(file.Name(), "v") {
			output = append(output, file.Name())
		}
	}

	return strings.Join(output, "\n"), nil
}
