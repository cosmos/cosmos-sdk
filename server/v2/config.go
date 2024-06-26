package serverv2

var ServerContextKey = struct{}{}

// Config is the config of the main server.
type Config struct {
	// StartBlock indicates if the server should block or not.
	// If true, the server will block until the context is canceled.
	StartBlock bool
}
