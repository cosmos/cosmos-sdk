package main

import (
	"cosmossdk.io/core/transaction"
	stf2 "cosmossdk.io/server/v2/stf"
)

type tx struct {
}

func (t tx) Hash() [32]byte {
	// TODO implement me
	panic("implement me")
}

func (t tx) GetMessages() ([]transaction.Type, error) {
	// TODO implement me
	panic("implement me")
}

func (t tx) GetSenders() ([]transaction.Identity, error) {
	// TODO implement me
	panic("implement me")
}

func (t tx) GetGasLimit() (uint64, error) {
	// TODO implement me
	panic("implement me")
}

func (t tx) Bytes() []byte {
	// TODO implement me
	panic("implement me")
}

func main() {
	stf := stf2.NewSTF[tx](
		nil,
		nil,
		nil, nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	stf.ValidateTx(nil, nil, 50, tx{})
}
