package staking

import crypto "github.com/tendermint/go-crypto"

type bondInfo struct {
	PubKey crypto.PubKey
	Power  int64
}

func (bi bondInfo) isEmpty() bool {
	if bi == (bondInfo{}) {
		return true
	}
	return false
}
