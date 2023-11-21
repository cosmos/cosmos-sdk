package transaction

type Codec[T Tx] interface {
	Encode(tx T) ([]byte, error)
	Decode([]byte) (T, error)
}
