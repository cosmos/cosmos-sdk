package module

// Option represents a module option.
type Option interface {
	todo()
}

// Provide registers providers with the dependency injection system. See
// github.com/cosmos/cosmos-sdk/container for more documentation on the
// dependency injection system.
func Provide(providers ...interface{}) Option {
	panic("TODO")
}
