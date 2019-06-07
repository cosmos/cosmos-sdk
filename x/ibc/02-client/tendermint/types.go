package tendermint

import (
	"bytes"

	"github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// Ref tendermint/lite/base_verifier.go

var _ client.ValidityPredicateBase = ValidityPredicateBase{}

type ValidityPredicateBase struct {
	Height           int64
	NextValidatorSet *types.ValidatorSet
}

func (ValidityPredicateBase) Kind() client.Kind {
	return client.Tendermint
}

func (base ValidityPredicateBase) GetHeight() int64 {
	return base.Height
}

func (base ValidityPredicateBase) Equal(cbase client.ValidityPredicateBase) bool {
	base0, ok := cbase.(ValidityPredicateBase)
	if !ok {
		return false
	}
	return base.Height == base0.Height &&
		bytes.Equal(base.NextValidatorSet.Hash(), base0.NextValidatorSet.Hash())
}

var _ client.Client = Client{}

type Client struct {
	Base ValidityPredicateBase
	Root commitment.Root
}

func (Client) Kind() client.Kind {
	return client.Tendermint
}

func (client Client) GetBase() client.ValidityPredicateBase {
	return client.Base
}

func (client Client) GetRoot() commitment.Root {
	return client.Root
}

func (client Client) Validate(header client.Header) (client.Client, error) {
	return client, nil // XXX
}

var _ client.Header = Header{}

type Header struct {
	Base  ValidityPredicateBase
	Root  commitment.Root
	Votes []*types.CommitSig
}

func (header Header) Kind() client.Kind {
	return client.Tendermint
}

func (header Header) GetBase() client.ValidityPredicateBase {
	return header.Base
}

func (header Header) GetRoot() commitment.Root {
	return header.Root
}
