package types

// FeeHandler optional custom fee handling implementations
type FeeHandler func(ctx Context, simulate bool) (coins Coins, events Events, err error)
