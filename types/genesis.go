package types

import cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"

type GenesisValidator struct {
	Address ConsAddress
	PubKey  cryptotypes.PubKey
	Power   int64
	Name    string
}
