package commitment

type State interface {
}

type Root interface {
}

type Proof interface {
	VerifyMembership(Root, []byte, []byte)
	VerifyNonMembership(Root, []byte)
}
