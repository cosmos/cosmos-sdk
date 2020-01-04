package types

type Marshaler interface {
	Marshal() (dAtA []byte, err error)
	MarshalTo([]byte) (int, error)
	Unmarshal(dAtA []byte) error
}
