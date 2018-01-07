package types

import crypto "github.com/tendermint/go-crypto"

type Model interface {
	Address() crypto.Address

	Get(key interface{}) interface{}
	Set(key interface{}, value interface{})
}

type ModelStore interface {
	Load(addr crypto.Address) Model
	Store(m Model)
}
