package keeper

import (
	"context"
	"errors"
	"fmt"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/event"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/consensus/exported"
	"github.com/cosmos/cosmos-sdk/x/consensus/types"
)

var StoreKey = "Consensus"

type Keeper struct {
	appmodule.Environment

	authority   string
	ParamsStore collections.Item[cmtproto.ConsensusParams]
	// storage of the last comet info
	cometInfo collections.Item[types.CometInfo]
}

var _ exported.ConsensusParamSetter = Keeper{}.ParamsStore

func NewKeeper(cdc codec.BinaryCodec, env appmodule.Environment, authority string) Keeper {
	sb := collections.NewSchemaBuilder(env.KVStoreService)
	return Keeper{
		Environment: env,
		authority:   authority,
		ParamsStore: collections.NewItem(sb, collections.NewPrefix("Consensus"), "params", codec.CollValue[cmtproto.ConsensusParams](cdc)),
	}
}

func (k *Keeper) GetAuthority() string {
	return k.authority
}

// Querier

var _ types.QueryServer = Keeper{}

// Params queries params of consensus module
func (k Keeper) Params(ctx context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	params, err := k.ParamsStore.Get(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryParamsResponse{Params: &params}, nil
}

// MsgServer

var _ types.MsgServer = Keeper{}

func (k Keeper) UpdateParams(ctx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if k.GetAuthority() != msg.Authority {
		return nil, fmt.Errorf("invalid authority; expected %s, got %s", k.GetAuthority(), msg.Authority)
	}

	consensusParams, err := msg.ToProtoConsensusParams()
	if err != nil {
		return nil, err
	}
	if err := cmttypes.ConsensusParamsFromProto(consensusParams).ValidateBasic(); err != nil {
		return nil, err
	}

	if err := k.ParamsStore.Set(ctx, consensusParams); err != nil {
		return nil, err
	}

	if err := k.EventService.EventManager(ctx).EmitKV(
		"update_consensus_params",
		event.NewAttribute("authority", msg.Authority),
		event.NewAttribute("parameters", consensusParams.String())); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

// SetParams sets the consensus parameters on init of a chain. This is a consensus message. It can only be called by the consensus server
// This is used in the consensus message handler set in module.go.
func (k Keeper) SetParams(ctx context.Context, req *types.ConsensusMsgParams) (*types.ConsensusMsgParamsResponse, error) {
	if req.Signer != "consensus" {
		return nil, fmt.Errorf("invalid signer; expected %s, got %s", "consensus", req.Signer)
	}

	consensusParams, err := req.ToProtoConsensusParams()
	if err != nil {
		return nil, err
	}
	if err := cmttypes.ConsensusParamsFromProto(consensusParams).ValidateBasic(); err != nil {
		return nil, err
	}

	if err := k.ParamsStore.Set(ctx, consensusParams); err != nil {
		return nil, err
	}

	return &types.ConsensusMsgParamsResponse{}, nil
}

// GetCometInfo returns info related to comet. If the application is using v1 then the information will be present on context,
// if the application is using v2 then the information will be present in the cometInfo store.
func (k Keeper) GetCometInfo(ctx context.Context, req *types.QueryCometInfoRequest) (*types.QueryCometInfoResponse, error) {
	ci, err := k.cometInfo.Get(ctx)
	// if the value is not found we may be using baseapp and not have consensus messages
	if errors.Is(err, collections.ErrNotFound) {
		ci := sdk.UnwrapSDKContext(ctx).CometInfo()
		res := &types.QueryCometInfoResponse{CometInfo: &types.CometInfo{
			ValidatorsHash:  ci.ValidatorsHash,
			ProposerAddress: ci.ProposerAddress,
		}}

		for _, vote := range ci.LastCommit.Votes {
			res.CometInfo.LastCommit.Votes = append(res.CometInfo.LastCommit.Votes, &types.VoteInfo{
				Validator: &types.Validator{
					Address: vote.Validator.Address,
					Power:   vote.Validator.Power,
				},
				BlockIdFlag: types.BlockIDFlag(vote.BlockIDFlag),
			})
		}
		res.CometInfo.LastCommit.Round = ci.LastCommit.Round
		for _, evi := range ci.Evidence {
			evi := evi
			res.CometInfo.Evidence = append(res.CometInfo.Evidence, &types.Evidence{
				EvidenceType: types.MisbehaviorType(evi.Type),
				Validator: &types.Validator{
					Address: evi.Validator.Address,
					Power:   evi.Validator.Power,
				},
				Height:           evi.Height,
				Time:             &evi.Time,
				TotalVotingPower: evi.TotalVotingPower,
			})
		}

		return res, err
	}

	return &types.QueryCometInfoResponse{CometInfo: &ci}, err
}

// SetCometInfo is called by the framework to set the value at genesis.
func (k Keeper) SetCometInfo(ctx context.Context, req *types.MsgCometInfoRequest) (*types.MsgCometInfoResponse, error) {
	if req.Signer != "consensus" { // TODO move this to core when up-streamed from server/v2
		return nil, fmt.Errorf("invalid signer; expected %s, got %s", "consensus", req.Signer)
	}

	err := k.cometInfo.Set(ctx, *req.CometInfo)

	return &types.MsgCometInfoResponse{}, err
}
