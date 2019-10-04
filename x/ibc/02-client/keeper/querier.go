package keeper

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/tendermint"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/merkle"
)

// NewQuerier creates a querier for the IBC client
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case types.QueryClientState:
			return queryClientState(ctx, req, k)
		case types.QueryConsensusState:
			return queryConsensusState(ctx)
		case types.QueryCommitmentPath:
			return queryCommitmentPath(ctx, req, k)
		case types.QueryCommitmentRoot:
			return queryCommitmentRoot(ctx, req, k)
		case types.QueryHeader:
			return queryHeader(ctx, req, k)
		default:
			return nil, sdk.ErrUnknownRequest("unknown IBC client query endpoint")
		}
	}
}

func queryClientState(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params types.QueryClientStateParams

	err := types.SubModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}


	// mapp := mapping(cdc, storeKey, version.Version)
	// state, _, err := k.State(id).ConsensusStateCLI(q)
	// if err != nil {
	// 	return err
	// }

	return res, nil
}

func queryConsensusState(ctx sdk.Context) ([]byte, sdk.Error) {

	state := tendermint.ConsensusState{
		ChainID:          commit.ChainID,
		Height:           uint64(commit.Height),
		Root:             merkle.NewRoot(commit.AppHash),
		NextValidatorSet: tmtypes.NewValidatorSet(validators.Validators),
	}

	return res, nil
}

func queryCommitmentPath(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {

	path := merkle.NewPrefix([][]byte{[]byte(k.mapping.storeName)}, k.mapping.PrefixBytes())

	return res, nil
}

func queryCommitmentRoot(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params types.QueryCommitmentRoot

	err := types.SubModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	root, _, err := k.ClientState(params.ID).RootCLI(q, params.Height)
	if err != nil {
		return nil, err
	}

	res, err := codec.MarshalJSONIndent(types.SubModuleCdc, root)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return res, nil
}

func queryHeader(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	// node, err := ctx.GetNode()
	// if err != nil {
	// 	return err
	// }

	// info, err := node.ABCIInfo()
	// if err != nil {
	// 	return err
	// }

	// height := info.Response.LastBlockHeight
	// prevheight := height - 1

	// commit, err := node.Commit(&height)
	// if err != nil {
	// 	return err
	// }

	// validators, err := node.Validators(&prevheight)
	// if err != nil {
	// 	return err
	// }

	// nextValidators, err := node.Validators(&height)
	// if err != nil {
	// 	return err
	// }

	// header := tendermint.Header{
	// 	SignedHeader:     commit.SignedHeader,
	// 	ValidatorSet:     tmtypes.NewValidatorSet(validators.Validators),
	// 	NextValidatorSet: tmtypes.NewValidatorSet(nextValidators.Validators),
	// }

	res, err := codec.MarshalJSONIndent(types.SubModuleCdc, header)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("failed to marshal result to JSON", err.Error()))
	}

	return res, nil
}
