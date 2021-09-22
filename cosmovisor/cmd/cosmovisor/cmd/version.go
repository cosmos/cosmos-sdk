package cmd

import (
	"fmt"
	"strings"
)

// Version represents Cosmovisor version value. Set during build
var Version string

func isVersionCommand(args []string) bool {
	return len(args) == 1 && strings.EqualFold(args[0], "version")
}

func printVersion() { fmt.Println("Cosmovisor Version: ", Version) }
