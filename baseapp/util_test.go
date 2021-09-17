package baseapp

import "github.com/cosmos/cosmos-sdk/types"

// CheckState is an exported method to be able to access baseapp's
// checkState in tests.
//
// This method is only accessible in baseapp tests.
func (app *BaseApp) CheckState() *state {
	return app.checkState
}

// DeliverState is an exported method to be able to access baseapp's
// deliverState in tests.
//
// This method is only accessible in baseapp tests.
func (app *BaseApp) DeliverState() *state {
	return app.deliverState
}

// CMS is an exported method to be able to access baseapp's cms in tests.
//
// This method is only accessible in baseapp tests.
func (app *BaseApp) CMS() types.CommitMultiStore {
	return app.cms
}
