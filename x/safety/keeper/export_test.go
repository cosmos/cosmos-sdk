package keeper

// NewMsgServerForTest exposes the unexported msg server to the keeper_test
// package.
var NewMsgServerForTest = newMsgServerImpl
