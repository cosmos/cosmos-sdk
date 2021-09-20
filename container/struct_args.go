package container

// StructArgs is a type which can be embedded in another struct to alert the
// container that the fields of the struct are dependency inputs/outputs. That
// is, the container will not look to resolve a value with StructArgs embedded
// directly, but will instead use the struct's fields to resolve or populate
// dependencies. Types with embedded StructArgs can be used in both the input
// and output parameter positions.
type StructArgs struct{}

func (StructArgs) isStructArgs() {}
