package store

// Used for associating paramstore key and field of param structs
type KeyValuePair struct {
	Key   []byte
	Value interface{}
}

// Slice of KeyFieldPair
type KeyValuePairs []KeyValuePair

// Interface for structs containing parameters for a module
type ParamStruct interface {
	KeyValuePairs() KeyValuePairs
}
