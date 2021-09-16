package cmd

// RunCosmovisorCommands executes cosmosvisor commands e.g `cosmovisor version`
func RunCosmovisorCommands(args []string) {
	if isVersionCommand(args) {
		printVersion()
	}
}
