package cmd

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Downloads the binary for the requested version if not found locally.",
	Long:  "Downloads the binary for the requested version if not found locally. Multiple versions can be downloaded at once: \"spin verman get 2.1.0 2.2.0\".",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("you must indicate the version of Spin you wish to set")
		}

		vermanDir, err := getVermanDir()
		if err != nil {
			return err
		}

		for _, version := range args {
			if !strings.HasPrefix(version, "v") {
				version = "v" + version
			}

			if err := downloadSpin(vermanDir, version); err != nil {
				return err
			}
		}

		return nil
	},
}

func downloadSpin(vermanDir, version string) error {
	var spinArch string
	var spinOS string

	// Checking for compatible architectures
	if runtime.GOARCH == "amd64" {
		spinArch = "amd64"
	} else if runtime.GOARCH == "arm64" {
		spinArch = "aarch64"
	} else {
		return fmt.Errorf("%q is not an architecture that Spin supports", runtime.GOARCH)
	}

	// Checking for compatible operating systems
	if runtime.GOOS == "linux" {
		// TODO: When would we want to download 'static-linux' vs just 'linux'?
		spinOS = "linux"
	} else if runtime.GOOS == "darwin" {
		spinOS = "macos"
	} else {
		return fmt.Errorf("%q is not an OS that this Spin plugin supports", runtime.GOOS)
	}

	fileDirectory := path.Join(vermanDir, "versions")
	fileName := fmt.Sprintf("spin-%s-%s-%s.tar.gz", version, spinOS, spinArch)

	dirExists, err := exists(fileDirectory)
	if err != nil {
		return err
	}

	// Determines if we need to pull the file from GitHub
	var versionFolderExists bool

	if !dirExists {
		if err = os.MkdirAll(fileDirectory, 0755); err != nil {
			return err
		}
	} else {
		dirFiles, err := os.ReadDir(fileDirectory)
		if err != nil {
			return err
		}

		for _, file := range dirFiles {
			// Checking if the Spin binary has previously been unpacked...
			if file.Name() == version {
				fmt.Printf("Spin version %s found locally.\n", version)
				versionFolderExists = true
				break
			}
		}
	}

	// If the tar.gz file doesn't exist, pull from GitHub
	if !versionFolderExists {
		fmt.Printf("Spin version %s not found locally. Retrieving from source...\n", version)

		resp, err := http.Get("https://github.com/fermyon/spin/releases/download/" + version + "/" + fileName)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("the version number provided is invalid: %s", version)
		}

		out, err := os.Create(path.Join(fileDirectory, fileName))
		if err != nil {
			return err
		}
		defer out.Close()

		_, err = io.Copy(out, resp.Body)
		if err != nil {
			return err
		}

		fmt.Printf("Spin version %s was retrieved successfully!\n", version)
		if err = unpackSpin(fileDirectory, fileName, version); err != nil {
			return err
		}
	}

	return nil
}

func unpackSpin(directory, tarGzFileName, version string) error {
	if err := os.Chdir(directory); err != nil {
		return err
	}

	gzipStream, err := os.ReadFile(tarGzFileName)
	if err != nil {
		return err
	}

	uncompressedStream, err := gzip.NewReader(bytes.NewReader(gzipStream))
	if err != nil {
		return err
	}

	tarReader := tar.NewReader(uncompressedStream)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return fmt.Errorf("unpackSpin: Next() failed: %w", err)
		}

		// Extracting only the Spin CLI binary
		if header.Typeflag == tar.TypeReg && header.Name == "spin" {
			// Create the file with the original permissions
			outFile, err := os.OpenFile(header.Name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			if _, err := io.Copy(outFile, tarReader); err != nil {
				return err
			}
			outFile.Close()

			// Ensure the file has the correct permissions
			if err := os.Chmod(header.Name, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("unpackSpin: could not set file permissions: %w", err)
			}
		}
	}

	// Create a folder named with the relevant Spin version
	if err := os.MkdirAll(version, 0755); err != nil {
		return err
	}

	if err := os.Rename("spin", path.Join(directory, version, "spin")); err != nil {
		return err
	}

	if err := os.Remove(tarGzFileName); err != nil {
		return err
	}

	return nil
}
