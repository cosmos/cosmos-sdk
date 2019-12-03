package subspace

type (
	// Param defines a parameter that any module may store or retrieve in a Subspace.
	// A Param is primarily composed of a key and a validation function that returns
	// an error if the Param value is considered invalid.
	Param interface {
		Key() []byte
		Validate() error
	}

	// SubKeyParam extends Param by requiring a subkey for a parameter.
	SubKeyParam interface {
		Param
		SubKey() []byte
	}

	// ParamSetPair is used for associating paramsubspace key and field of param
	// structs.
	ParamSetPair struct {
		Param Param
		Value interface{}
	}
)

// NewParamSetPair creates a new ParamSetPair instance
func NewParamSetPair(param Param, value interface{}) ParamSetPair {
	return ParamSetPair{param, value}
}

// ParamSetPairs Slice of KeyFieldPair
type ParamSetPairs []ParamSetPair

// ParamSet defines an interface for structs containing parameters for a module
type ParamSet interface {
	ParamSetPairs() ParamSetPairs
}

var _ Param = (*ephemeralParam)(nil)

// ephemeralParam defines a Param type used for common Subspace functionality
// across Param and SubKeyParam types so a common API can be used.
type ephemeralParam struct {
	key []byte
}

func NewEphemeralParam(key, subkey []byte) Param { return ephemeralParam{key: concatKeys(key, subkey)} }
func (ep ephemeralParam) Key() []byte            { return ep.key }
func (ep ephemeralParam) Validate() error        { return nil }
