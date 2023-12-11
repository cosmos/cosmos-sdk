package transaction

type Codec[T Tx] interface {
	Decode([]byte) (T, error)
}
