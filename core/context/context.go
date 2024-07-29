package context

type (
	execModeKey    struct{}
	cometInfoKey   struct{}
	environmentKey struct{}
)

var (
	ExecModeKey  = execModeKey{}
	CometInfoKey = cometInfoKey{}

	// EnvironmentContextKey is the context key for the environment.
	// A caller should not assume the environment is available in each context.
	// ref: https://github.com/cosmos/cosmos-sdk/issues/19640
	EnvironmentContextKey = environmentKey{}
)
