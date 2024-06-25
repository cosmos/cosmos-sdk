package helpers

import (
	"os"
	"path/filepath"
	"strings"
)

// GetNodeHomeDirectory gets the home directory of the node (where the config is located).
// It parses the home flag if set if the `NODE_HOME` environment variable if set (and ignores name).
// Otherwise, it returns the default home directory given its name.
func GetNodeHomeDirectory(name string) (string, error) {
	// get the home directory from the flag
	args := os.Args
	for i := 0; i < len(args); i++ {
		if args[i] == "--home" && i+1 < len(args) {
			return filepath.Clean(args[i+1]), nil
		} else if strings.HasPrefix(args[i], "--home=") {
			return filepath.Clean(args[i][7:]), nil
		}
	}

	// get the home directory from the environment variable
	homeDir := os.Getenv("NODE_HOME")
	if homeDir != "" {
		return filepath.Clean(homeDir), nil
	}

	// return the default home directory
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(userHomeDir, name), nil
}
