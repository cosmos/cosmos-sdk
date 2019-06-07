package commitment

// XXX: []byte?
type Root interface{}

// XXX: need to separate membership and non membership proof types
type Proof interface {
	Key() []byte
	Verify(Root, []byte, []byte) error
}

type FullProof struct {
	Proof Proof
	Value []byte
}

func (proof FullProof) Verify(root Root) error {
	return proof.Proof.Verify(root, proof.Proof.Key(), proof.Value)
}
