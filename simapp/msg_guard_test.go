package simapp

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil/msgguard"
)

// TestMsgServersCheckCircuitBreakerFirst calls every msg server method SimApp
// registers directly, as an EVM precompile calls them, and requires each one
// to check the circuit breaker before anything else. Only the msgs x/safety
// always allows may skip the check. A new msg server method without the guard
// fails here.
func TestMsgServersCheckCircuitBreakerFirst(t *testing.T) {
	// The sweep needs only the module manager and router, not a chain.
	app, _ := setup(false, 0)
	msgguard.RequireAppGuarded(t, app.AppCodec(), app.ModuleManager, app.InterfaceRegistry(), app.MsgServiceRouter())
}
