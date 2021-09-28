package cmd

// RunCosmovisorCommands executes cosmosvisor commands e.g `cosmovisor version`
// Returned boolean is whether or not execution should continue.
func RunCosmovisorCommands(args []string) {
	switch {
	case ShouldGiveHelp(args):
		DoHelp()
	case isVersionCommand(args):
		printVersion()
	}
}
