package utils

import (
	"fmt"

	"github.com/spf13/viper"

	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/tendermint"
)

// QueryConsensusStateProof queries the store to get the consensus state and a
// merkle proof.
func QueryConsensusStateProof(
	cliCtx client.CLIContext, cdc *codec.Codec,
	queryRoute, clientID string,
) (types.ConsensusStateResponse, error) {
	var conStateRes types.ConsensusStateResponse

	bz, err := cdc.MarshalJSON(types.NewQueryClientStateParams(clientID))
	if err != nil {
		return conStateRes, err
	}

	req := abci.RequestQuery{
		Path:  fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryClientState),
		Data:  bz,
		Prove: viper.GetBool(flags.FlagProve),
	}

	res, err := cliCtx.QueryABCI(req)
	if err != nil {
		return conStateRes, err
	}

	var cs tendermint.ConsensusState
	if err := cliCtx.Codec.UnmarshalJSON(res.Value, &cs); err != nil {
		return conStateRes, err
	}

	return types.NewConsensusStateResponse(clientID, cs, res.Proof, res.Height), nil
}

// GetTendermintHeader takes a client context and returns the appropriate
// tendermint header
func GetTendermintHeader(cliCtx context.CLIContext) (tendermint.Header, error) {
	node, err := cliCtx.GetNode()
	if err != nil {
		return tendermint.Header{}, err
	}

	info, err := node.ABCIInfo()
	if err != nil {
		return tendermint.Header{}, err
	}

	height := info.Response.LastBlockHeight
	prevheight := height - 1

	commit, err := node.Commit(&height)
	if err != nil {
		return tendermint.Header{}, err
	}

	validators, err := node.Validators(&prevheight)
	if err != nil {
		return tendermint.Header{}, err
	}

	nextvalidators, err := node.Validators(&height)
	if err != nil {
		return tendermint.Header{}, err
	}

	header := tendermint.Header{
		SignedHeader:     commit.SignedHeader,
		ValidatorSet:     tmtypes.NewValidatorSet(validators.Validators),
		NextValidatorSet: tmtypes.NewValidatorSet(nextvalidators.Validators),
	}

	return header, nil
}
