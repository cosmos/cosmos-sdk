package main

import (
	"os"

	confixcmd "cosmossdk.io/tools/confix/cmd"
)

func main() {
	if err := confixcmd.ConfixCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
