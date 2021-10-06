package cmd

import (
	"fmt"
)

// Version represents Cosmovisor version value. Set during build
var Version string

// VersionArgs is the strings that indicate a cosmovisor version command.
var VersionArgs = []string{"version", "--version"}

// IsVersionCommand checks if the given args indicate that the version is being requested.
func IsVersionCommand(arg string) bool {
	return isOneOf(arg, VersionArgs)
}

// PrintVersion prints the cosmovisor version.
func PrintVersion() {
	fmt.Println("Cosmovisor Version: ", Version)
}
