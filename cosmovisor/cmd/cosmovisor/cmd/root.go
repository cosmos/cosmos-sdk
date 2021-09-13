package cmd

func RunCosmovisorCommands(args []string) {

	if isVersionCommand(args) {
		printVersion()
	}

}
