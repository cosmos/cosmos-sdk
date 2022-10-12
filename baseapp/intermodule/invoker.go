package intermodule

import "cosmossdk.io/core/appmodule"

type InvokerFactory func(callInfo CallInfo) (appmodule.InterModuleInvoker, error)

type CallInfo struct {
	Method      string
	DerivedPath []byte
}
