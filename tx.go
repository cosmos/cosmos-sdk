package basecoin

// TODO: add some common functionality here...
// +gen wrapper:"Tx"
type TxInner interface {
	Wrap() Tx
}

// TxLayer provides a standard way to deal with "middleware" tx,
// That add context to an embedded tx.
type TxLayer interface {
	TxInner
	Next() Tx
}

func (t Tx) IsLayer() bool {
	_, ok := t.Unwrap().(TxLayer)
	return ok
}

func (t Tx) GetLayer() TxLayer {
	l, _ := t.Unwrap().(TxLayer)
	return l
}
