package tendermint

import (
	"bytes"
	"errors"
	"fmt"

	lerr "github.com/tendermint/tendermint/lite/errors"
	"github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

var _ client.ConsensusState = ConsensusState{}

// Ref tendermint/lite/base_verifier.go
type ConsensusState struct {
	ChainID          string
	Height           uint64
	Root             commitment.Root
	NextValidatorSet *types.ValidatorSet
}

func (ConsensusState) Kind() client.Kind {
	return client.Tendermint
}

func (cs ConsensusState) GetHeight() uint64 {
	return cs.Height
}

func (cs ConsensusState) GetRoot() commitment.Root {
	return cs.Root
}

func (cs ConsensusState) update(header Header) ConsensusState {
	return ConsensusState{
		ChainID:          cs.ChainID,
		Height:           uint64(header.Height),
		Root:             header.AppHash,
		NextValidatorSet: header.NextValidatorSet,
	}
}

func (cs ConsensusState) Validate(cheader client.Header) (client.ConsensusState, error) {
	header, ok := cheader.(Header)
	if !ok {
		return nil, errors.New("invalid type")
	}

	nextvalset := cs.NextValidatorSet
	nexthash := nextvalset.Hash()

	if cs.Height == uint64(header.Height-1) {
		nexthash = cs.NextValidatorSet.Hash()
		if !bytes.Equal(header.ValidatorsHash, nexthash) {
			fmt.Println(111)
			return nil, lerr.ErrUnexpectedValidators(header.ValidatorsHash, nexthash)
		}
	}

	if !bytes.Equal(header.NextValidatorsHash, header.NextValidatorSet.Hash()) {
		fmt.Println(header)
		return nil, lerr.ErrUnexpectedValidators(header.NextValidatorsHash, header.NextValidatorSet.Hash())
	}

	err := header.ValidateBasic(cs.ChainID)
	if err != nil {
		return nil, err
	}

	err = cs.NextValidatorSet.VerifyCommit(cs.ChainID, header.Commit.BlockID, header.Height, header.Commit)
	if err != nil {
		return nil, err
	}

	return cs.update(header), nil
}

func (cs ConsensusState) Equivocation(header1, header2 client.Header) bool {
	return false // XXX
}

var _ client.Header = Header{}

type Header struct {
	// XXX: don't take the entire struct
	types.SignedHeader
	NextValidatorSet *types.ValidatorSet
}

func (header Header) Kind() client.Kind {
	return client.Tendermint
}

func (header Header) GetHeight() uint64 {
	return uint64(header.Height)
}
