package subspace

// ParamSetPair is used for associating paramsubspace key and field of param structs
type ParamSetPair struct {
	Key   []byte
	Value interface{}
}

// NewParamSetPair creates a new ParamSetPair instance
func NewParamSetPair(key []byte, value interface{}) ParamSetPair {
	return ParamSetPair{key, value}
}

// ParamSetPairs Slice of KeyFieldPair
type ParamSetPairs []ParamSetPair

// ParamSet defines an interface for structs containing parameters for a module
type ParamSet interface {
	ParamSetPairs() ParamSetPairs
}
