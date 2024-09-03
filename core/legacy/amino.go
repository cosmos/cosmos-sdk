package legacy

// Amino is an interface that allow to register concrete types and interfaces with the Amino codec.
type Amino interface {
	// RegisterInterface registers an interface and its concrete type with the Amino codec.
	RegisterInterface(interfacePtr any, iopts *InterfaceOptions)

	// RegisterConcrete registers a concrete type with the Amino codec.
	RegisterConcrete(cdcType interface{}, name string)
}

// InterfaceOptions defines options for registering an interface with the Amino codec.
type InterfaceOptions struct {
	Priority           []string // Disamb priority.
	AlwaysDisambiguate bool     // If true, include disamb for all types.
}
