package context

type contextKey uint8

const (
	ExecModeKey  contextKey = iota
	CometInfoKey contextKey = iota
)
