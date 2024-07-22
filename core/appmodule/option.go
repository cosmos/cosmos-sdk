package appmodule

import (
	depinjectappconfig "cosmossdk.io/depinject/appconfig"
)

// Option is a functional option for implementing modules.
type Option = depinjectappconfig.Option

// Provide registers providers with the dependency injection system that will be
// run within the module scope. See cosmossdk.io/depinject for
// documentation on the dependency injection system.
var Provide = depinjectappconfig.Provide

// Invoke registers invokers to run with depinject. Each invoker will be called
// at the end of dependency graph configuration in the order in which it was defined. Invokers may not define output
// parameters, although they may return an error, and all of their input parameters will be marked as optional so that
// invokers impose no additional constraints on the dependency graph. Invoker functions should nil-check all inputs.
var Invoke = depinjectappconfig.Invoke
