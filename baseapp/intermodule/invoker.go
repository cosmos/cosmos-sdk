package intermodule

import "cosmossdk.io/core/appmodule"

type InvokerFactory func(callInfo CallInfo) (appmodule.Invoker, error)

type CallInfo struct {
	Method      string
	DerivedPath []byte
}
