package store

// Used for associating paramstore key and field of param structs
type KeyFieldPair struct {
	Key   []byte
	Field interface{}
}

// Slice of KeyFieldPair
type KeyValuePairs []KeyFieldPair

// Interface for structs containing parameters for a module
type ParamStruct interface {
	KeyValuePairs() KeyValuePairs
}
