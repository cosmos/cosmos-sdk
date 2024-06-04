package context

type contextKey uint8

const (
	ExecModeKey  contextKey = iota
	CometInfoKey contextKey = iota
)

// EnvironmentContextKey is the context key for the environment.
// A caller should not assume the environment is available in each context.
// ref: https://github.com/cosmos/cosmos-sdk/issues/19640
var EnvironmentContextKey = struct{}{}
