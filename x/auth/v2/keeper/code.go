package keeper

type CodeType uint8

const (
	// CodeTypePubKey represents a public key.
	CodeTypePubKey CodeType = iota
	// CodeTypeVM represents a VM abstract account.
	CodeTypeVM
	// CodeTypeModule represents a module account.
	CodeTypeModule
)

// CodeData represents either a public key or a pointer to an account's implementation code in some VM or module.
type CodeData struct {
	// Type represents the type of code.
	Type CodeType
	// Handler represents the name of the public key algorithm, VM or module.
	Handler string
	// Data represents the public key, a VM code identifier, or a module-specific identifier.
	Data []byte
}
