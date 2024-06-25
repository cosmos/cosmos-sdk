package helpers

import (
	"os"
	"path/filepath"
	"strings"
)

// GetNodeHomeDirectory gets the home directory of the node (where the config is located)
// It parses the home flag if set or an environment variable if set
// Otherwise, it returns the default home directory
func GetNodeHomeDirectory(name string) (string, error) {
	// get the home directory from the flag
	args := os.Args
	for i := 0; i < len(args); i++ {
		if args[i] == "--home" && i+1 < len(args) {
			return filepath.Join(args[i+1], name), nil
		} else if strings.HasPrefix(args[i], "--home=") {
			return filepath.Join(args[i][7:], name), nil
		}
	}

	// get the home directory from the environment variable
	homeDir := os.Getenv("HOME")
	if homeDir != "" {
		return filepath.Join(homeDir, name), nil
	}

	// return the default home directory
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(userHomeDir, name), nil
}
