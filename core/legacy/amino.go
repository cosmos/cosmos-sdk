package legacy

type Amino interface {
	// RegisterInterface registers an interface and its concrete type with the Amino codec.
	RegisterInterface(interfacePtr, cdcType interface{})

	// RegisterConcrete registers a concrete type with the Amino codec.
	RegisterConcrete(cdcType interface{}, name string)
}
