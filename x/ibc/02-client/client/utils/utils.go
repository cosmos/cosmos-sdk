package utils

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// QueryAllClientStates returns all the light client states. It _does not_ return
// any merkle proof.
func QueryAllClientStates(clientCtx client.Context, page, limit int) ([]exported.ClientState, int64, error) {
	params := types.NewQueryAllClientsParams(page, limit)
	bz, err := clientCtx.JSONMarshaler.MarshalJSON(params)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to marshal query params: %w", err)
	}

	route := fmt.Sprintf("custom/%s/%s/%s", "ibc", types.QuerierRoute, types.QueryAllClients)
	res, height, err := clientCtx.QueryWithData(route, bz)
	if err != nil {
		return nil, 0, err
	}

	var clients []exported.ClientState
	err = clientCtx.JSONMarshaler.UnmarshalJSON(res, &clients)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to unmarshal light clients: %w", err)
	}
	return clients, height, nil
}

// QueryClientState queries the store to get the light client state and a merkle
// proof.
func QueryClientState(
	clientCtx client.Context, clientID string, prove bool,
) (types.StateResponse, error) {
	req := abci.RequestQuery{
		Path:  "store/ibc/key",
		Data:  host.FullKeyClientPath(clientID, host.KeyClientState()),
		Prove: prove,
	}

	res, err := clientCtx.QueryABCI(req)
	if err != nil {
		return types.StateResponse{}, err
	}

	var clientState exported.ClientState
	if err := clientCtx.Codec.UnmarshalBinaryBare(res.Value, &clientState); err != nil {
		return types.StateResponse{}, err
	}

	clientStateRes := types.NewClientStateResponse(clientID, clientState, res.ProofOps, res.Height)

	return clientStateRes, nil
}

// QueryConsensusState queries the store to get the consensus state and a merkle
// proof.
func QueryConsensusState(
	clientCtx client.Context, clientID string, height uint64, prove bool,
) (types.ConsensusStateResponse, error) {
	var conStateRes types.ConsensusStateResponse

	req := abci.RequestQuery{
		Path:  "store/ibc/key",
		Data:  host.FullKeyClientPath(clientID, host.KeyConsensusState(height)),
		Prove: prove,
	}

	res, err := clientCtx.QueryABCI(req)
	if err != nil {
		return conStateRes, err
	}

	var cs exported.ConsensusState
	if err := clientCtx.Codec.UnmarshalBinaryBare(res.Value, &cs); err != nil {
		return conStateRes, err
	}

	return types.NewConsensusStateResponse(clientID, cs, res.ProofOps, res.Height), nil
}

// QueryTendermintHeader takes a client context and returns the appropriate
// tendermint header
func QueryTendermintHeader(clientCtx client.Context) (ibctmtypes.Header, int64, error) {
	node, err := clientCtx.GetNode()
	if err != nil {
		return ibctmtypes.Header{}, 0, err
	}

	info, err := node.ABCIInfo()
	if err != nil {
		return ibctmtypes.Header{}, 0, err
	}

	height := info.Response.LastBlockHeight

	commit, err := node.Commit(&height)
	if err != nil {
		return ibctmtypes.Header{}, 0, err
	}

	page := 0
	count := 10_000

	validators, err := node.Validators(&height, &page, &count)
	if err != nil {
		return ibctmtypes.Header{}, 0, err
	}

	header := ibctmtypes.Header{
		SignedHeader: commit.SignedHeader,
		ValidatorSet: tmtypes.NewValidatorSet(validators.Validators),
	}

	return header, height, nil
}

// QueryNodeConsensusState takes a client context and returns the appropriate
// tendermint consensus state
func QueryNodeConsensusState(clientCtx client.Context) (*ibctmtypes.ConsensusState, int64, error) {
	node, err := clientCtx.GetNode()
	if err != nil {
		return &ibctmtypes.ConsensusState{}, 0, err
	}

	info, err := node.ABCIInfo()
	if err != nil {
		return &ibctmtypes.ConsensusState{}, 0, err
	}

	height := info.Response.LastBlockHeight

	commit, err := node.Commit(&height)
	if err != nil {
		return &ibctmtypes.ConsensusState{}, 0, err
	}

	page := 0
	count := 10_000

	nextHeight := height + 1
	nextVals, err := node.Validators(&nextHeight, &page, &count)
	if err != nil {
		return &ibctmtypes.ConsensusState{}, 0, err
	}

	state := &ibctmtypes.ConsensusState{
		Timestamp:          commit.Time,
		Root:               commitmenttypes.NewMerkleRoot(commit.AppHash),
		NextValidatorsHash: tmtypes.NewValidatorSet(nextVals.Validators).Hash(),
	}

	return state, height, nil
}
