package utils

import (
	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/tendermint"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// QueryClientState queries the store to get the light client state and a merkle
// proof.
func QueryClientState(
	cliCtx context.CLIContext, clientID string, prove bool,
) (types.StateResponse, error) {
	req := abci.RequestQuery{
		Path:  "store/ibc/key",
		Data:  types.KeyClientState(clientID),
		Prove: prove,
	}

	res, err := cliCtx.QueryABCI(req)
	if err != nil {
		return types.StateResponse{}, err
	}

	var clientState types.State
	if err := cliCtx.Codec.UnmarshalJSON(res.Value, &clientState); err != nil {
		return types.StateResponse{}, err
	}

	clientStateRes := types.NewClientStateResponse(clientID, clientState, res.Proof, res.Height)

	return clientStateRes, nil
}

// QueryConsensusStateProof queries the store to get the consensus state and a
// merkle proof.
func QueryConsensusStateProof(
	cliCtx client.CLIContext, clientID string, prove bool) (types.ConsensusStateResponse, error) {
	var conStateRes types.ConsensusStateResponse

	req := abci.RequestQuery{
		Path:  "store/ibc/key",
		Data:  types.KeyConsensusState(clientID),
		Prove: prove,
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

// QueryCommitmentRoot queries the store to get the commitment root and a merkle
// proof.
func QueryCommitmentRoot(
	cliCtx context.CLIContext, clientID string, height uint64, prove bool,
) (types.RootResponse, error) {
	req := abci.RequestQuery{
		Path:  "store/ibc/key",
		Data:  types.KeyRoot(clientID, height),
		Prove: prove,
	}

	res, err := cliCtx.QueryABCI(req)
	if err != nil {
		return types.RootResponse{}, err
	}

	var root commitment.Root
	if err := cliCtx.Codec.UnmarshalJSON(res.Value, &root); err != nil {
		return types.RootResponse{}, err
	}

	rootRes := types.NewRootResponse(clientID, height, root, res.Proof, res.Height)

	return rootRes, nil
}

// QueryTendermintHeader takes a client context and returns the appropriate
// tendermint header
func QueryTendermintHeader(cliCtx context.CLIContext) (tendermint.Header, int64, error) {
	node, err := cliCtx.GetNode()
	if err != nil {
		return tendermint.Header{}, 0, err
	}

	info, err := node.ABCIInfo()
	if err != nil {
		return tendermint.Header{}, 0, err
	}

	height := info.Response.LastBlockHeight
	prevheight := height - 1

	commit, err := node.Commit(&height)
	if err != nil {
		return tendermint.Header{}, 0, err
	}

	validators, err := node.Validators(&prevheight)
	if err != nil {
		return tendermint.Header{}, 0, err
	}

	nextvalidators, err := node.Validators(&height)
	if err != nil {
		return tendermint.Header{}, 0, err
	}

	header := tendermint.Header{
		SignedHeader:     commit.SignedHeader,
		ValidatorSet:     tmtypes.NewValidatorSet(validators.Validators),
		NextValidatorSet: tmtypes.NewValidatorSet(nextvalidators.Validators),
	}

	return header, height, nil
}

// QueryNodeConsensusState takes a client context and returns the appropriate
// tendermint consensus state
func QueryNodeConsensusState(cliCtx context.CLIContext) (tendermint.ConsensusState, int64, error) {
	node, err := cliCtx.GetNode()
	if err != nil {
		return tendermint.ConsensusState{}, 0, err
	}

	info, err := node.ABCIInfo()
	if err != nil {
		return tendermint.ConsensusState{}, 0, err
	}

	height := info.Response.LastBlockHeight
	prevHeight := height - 1

	commit, err := node.Commit(&height)
	if err != nil {
		return tendermint.ConsensusState{}, 0, err
	}

	validators, err := node.Validators(&prevHeight)
	if err != nil {
		return tendermint.ConsensusState{}, 0, err
	}

	state := tendermint.ConsensusState{
		ChainID:          commit.ChainID,
		Height:           uint64(commit.Height),
		Root:             commitment.NewRoot(commit.AppHash),
		NextValidatorSet: tmtypes.NewValidatorSet(validators.Validators),
	}

	return state, height, nil
}
