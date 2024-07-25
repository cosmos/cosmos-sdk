package types

// Context is an interface used by an App to pass context information
// needed to process store streaming requests.
type Context interface {
	BlockHeight() int64
	Logger() Logger
	StreamingManager() StreamingManager
}
