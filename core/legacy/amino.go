package legacy

type Amino interface {
	// RegisterInterface registers an interface and its concrete type with the Amino codec.
	RegisterInterface(interfacePtr any, iopts *InterfaceOptions)

	// RegisterConcrete registers a concrete type with the Amino codec.
	RegisterConcrete(cdcType interface{}, name string)
}

type InterfaceOptions struct {
	Priority           []string // Disamb priority.
	AlwaysDisambiguate bool     // If true, include disamb for all types.
}
