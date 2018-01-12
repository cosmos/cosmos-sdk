package types

// A generic codec for a fixed type.
type Codec interface {

	// Returns a prototype (empty) object.
	Prototype() interface{}

	// Encodes the object.
	Encode(o interface{}) ([]byte, error)

	// Decodes an object.
	Decode(bz []byte) (interface{}, error)
}
