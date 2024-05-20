package context

type ServerContextKey string

const (
	LoggerContextKey ServerContextKey = "server.logger"
	ViperContextKey  ServerContextKey = "server.viper"
)
