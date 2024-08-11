package context

type (
	execModeKey    struct{}
	cometInfoKey   struct{}
	initInfoKey    struct{}
	environmentKey struct{}
)

var (
	ExecModeKey  = execModeKey{}
	CometInfoKey = cometInfoKey{}
	InitInfoKey  = initInfoKey{}

	// EnvironmentContextKey is the context key for the environment.
	// A caller should not assume the environment is available in each context.
	// ref: https://github.com/cosmos/cosmos-sdk/issues/19640
	EnvironmentContextKey = environmentKey{}
)
