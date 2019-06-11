package commitment

// XXX: []byte?
type Root interface{}

// XXX: need to separate membership and non membership proof types
type Proof interface {
	Key() []byte
	Verify(Root, []byte) error
}
