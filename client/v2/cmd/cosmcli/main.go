package main

import "cosmossdk.io/client/v2/autocli"

func main() {
	cmd, err := autocli.RemoteCommandOptions{}.Command()
	if err != nil {
		panic(err)
	}

	err = cmd.Execute()
	if err != nil {
		panic(err)
	}
}
