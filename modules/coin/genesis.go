package coin

import (
	"bytes"

	"github.com/pkg/errors"

	crypto "github.com/tendermint/go-crypto"
	"github.com/tendermint/go-wire/data"

	"github.com/tendermint/basecoin/types"
)

/**** code to parse accounts from genesis docs ***/

// GenesisAccount - genesis account parameters
type GenesisAccount struct {
	Address data.Bytes `json:"address"`
	// this from types.Account (don't know how to embed this properly)
	PubKey   crypto.PubKey `json:"pub_key"` // May be nil, if not known.
	Sequence int           `json:"sequence"`
	Balance  types.Coins   `json:"coins"`
}

// ToAccount - GenesisAccount struct to a basecoin Account
func (g GenesisAccount) ToAccount() Account {
	return Account{
		Sequence: g.Sequence,
		Coins:    g.Balance,
	}
}

// GetAddr - Get the address of the genesis account
func (g GenesisAccount) GetAddr() ([]byte, error) {
	noAddr, noPk := len(g.Address) == 0, g.PubKey.Empty()

	if noAddr {
		if noPk {
			return nil, errors.New("No address given")
		}
		return g.PubKey.Address(), nil
	}
	if noPk { // but is addr...
		return g.Address, nil
	}
	// now, we have both, make sure they check out
	if bytes.Equal(g.Address, g.PubKey.Address()) {
		return g.Address, nil
	}
	return nil, errors.New("Address and pubkey don't match")
}
