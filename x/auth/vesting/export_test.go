package vesting

// NewMsgServerForTest exposes the unexported msg server to the vesting_test
// package.
var NewMsgServerForTest = newMsgServerImpl
