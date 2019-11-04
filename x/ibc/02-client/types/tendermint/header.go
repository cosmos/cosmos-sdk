package tendermint

import (
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
)

var _ exported.Header = Header{}

// Header defines the Tendermint consensus Header
type Header struct {
	tmtypes.SignedHeader
	ValidatorSet     *tmtypes.ValidatorSet `json:"validator_set" yaml:"validator_set"`
	NextValidatorSet *tmtypes.ValidatorSet `json:"next_validator_set" yaml:"next_validator_set"`
}

// ClientType defines that the Header is a Tendermint consensus algorithm
func (h Header) ClientType() exported.ClientType {
	return exported.Tendermint
}

// GetHeight returns the current height
//
// NOTE: also referred as `sequence`
func (h Header) GetHeight() uint64 {
	return uint64(h.Height)
}

// GetHeader takes a context and returns the appropriate header
func GetHeader(ctx context.CLIContext) (Header, error) {
	node, err := ctx.GetNode()
	if err != nil {
		return Header{}, err
	}

	info, err := node.ABCIInfo()
	if err != nil {
		return Header{}, err
	}

	height := info.Response.LastBlockHeight
	prevheight := height - 1

	commit, err := node.Commit(&height)
	if err != nil {
		return Header{}, err
	}

	validators, err := node.Validators(&prevheight)
	if err != nil {
		return Header{}, err
	}

	nextvalidators, err := node.Validators(&height)
	if err != nil {
		return Header{}, err
	}

	header := Header{
		SignedHeader:     commit.SignedHeader,
		ValidatorSet:     tmtypes.NewValidatorSet(validators.Validators),
		NextValidatorSet: tmtypes.NewValidatorSet(nextvalidators.Validators),
	}

	return header, nil
}
