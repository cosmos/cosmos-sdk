package commitment

// XXX: []byte?
type Root interface{}

// XXX: need to separate membership and non membership proof types
type Proof interface {
	Verify(Root, []byte, []byte) error
}

type FullProof struct {
	Proof Proof
	Key   []byte
	Value []byte
}

func (proof FullProof) Verify(root Root) error {
	return proof.Proof.Verify(root, proof.Key, proof.Value)
}
