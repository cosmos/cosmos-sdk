//go:build !consensus_break_test

package baseapp

// injectConsensusBreak is a dummy function for production builds where the
// consensus_break_test build tag is not included. It does nothing.
func (app *BaseApp) injectConsensusBreak(appHash []byte, height int64, path string) []byte {
	return appHash
}
