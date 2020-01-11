package proto

type Codec struct{}

type Marshaler interface {
	Marshal() ([]byte, error)
	MarshalTo(data []byte) (n int, err error)
	MarshalToSizedBuffer(dAtA []byte) (int, error)
	Size() int
	Unmarshal(data []byte) error
}
