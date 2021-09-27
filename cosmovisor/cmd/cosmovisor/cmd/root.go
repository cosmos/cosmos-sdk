package cmd

// RunCosmovisorCommands executes cosmosvisor commands e.g `cosmovisor version`
// Returned boolean is whether or not execution should continue.
func RunCosmovisorCommands(args []string) bool {
	switch {
	case ShouldGiveHelp(args):
		DoHelp(args)
		return false
	case isVersionCommand(args):
		printVersion()
	}
	return true
}
