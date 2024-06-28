package context

type (
	loggerContextKey struct{}
	viperContextKey  struct{}
)

var (
	LoggerContextKey loggerContextKey
	ViperContextKey  viperContextKey
)
