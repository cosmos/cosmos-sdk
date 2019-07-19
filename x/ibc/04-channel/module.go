package channel

type IBCModule interface {
	NewIBCHandler() Handler
	Name() string
}
