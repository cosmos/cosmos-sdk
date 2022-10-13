package main

import "cosmossdk.io/client/v2/cli"

func main() {
	cmd, err := cli.RemoteCommand(cli.RemoteCommandOptions{ConfigDir: "TODO"})
	if err != nil {
		panic(err)
	}

	err = cmd.Execute()
	if err != nil {
		panic(err)
	}
}
