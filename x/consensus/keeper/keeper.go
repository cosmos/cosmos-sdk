package keeper

import (
	"context"
	"fmt"

	abci "github.com/cometbft/cometbft/api/cometbft/abci/v1"
	cmtproto "github.com/cometbft/cometbft/api/cometbft/types/v1"
	cmttypes "github.com/cometbft/cometbft/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/comet"
	"cosmossdk.io/core/event"
	"cosmossdk.io/x/consensus/exported"
	"cosmossdk.io/x/consensus/types"

	"github.com/cosmos/cosmos-sdk/codec"
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
		cometInfo:   collections.NewItem(sb, collections.NewPrefix("CometInfo"), "comet_info", codec.CollValue[types.CometInfo](cdc)),
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

// UpdateCometInfo implements types.MsgServer.
func (k Keeper) UpdateCometInfo(ctx context.Context, req *types.MsgUpdateCometInfo) (*types.MsgUpdateCometInfoResponse, error) {
	if k.GetAuthority() != req.Authority {
		return nil, fmt.Errorf("invalid signer; expected %s, got %s", k.GetAuthority(), req.Authority)
	}

	err := k.cometInfo.Set(ctx, *req.CometInfo)

	return &types.MsgUpdateCometInfoResponse{}, err
}

func (k Keeper) GetCometInfo(ctx context.Context) (*comet.Info, error) {
	ci, err := k.cometInfo.Get(ctx)
	if err != nil {
		return nil, err
	}
	res := &comet.Info{
		ProposerAddress: ci.ProposerAddress,
		ValidatorsHash:  ci.ValidatorsHash,
		Evidence:        toCoreEvidence(ci.Evidence),
		LastCommit:      toCoreCommitInfo(ci.LastCommit),
	}
	return res, nil
}

// toCoreEvidence takes comet evidence and returns sdk evidence
func toCoreEvidence(ev []abci.Misbehavior) []comet.Evidence {
	evidence := make([]comet.Evidence, len(ev))
	for i, e := range ev {
		evidence[i] = comet.Evidence{
			Type:             comet.MisbehaviorType(e.Type),
			Height:           e.Height,
			Time:             e.Time,
			TotalVotingPower: e.TotalVotingPower,
			Validator: comet.Validator{
				Address: e.Validator.Address,
				Power:   e.Validator.Power,
			},
		}
	}
	return evidence
}

// toCoreCommitInfo takes comet commit info and returns sdk commit info
func toCoreCommitInfo(commit abci.CommitInfo) comet.CommitInfo {
	ci := comet.CommitInfo{
		Round: commit.Round,
	}

	for _, v := range commit.Votes {
		ci.Votes = append(ci.Votes, comet.VoteInfo{
			Validator: comet.Validator{
				Address: v.Validator.Address,
				Power:   v.Validator.Power,
			},
			BlockIDFlag: comet.BlockIDFlag(v.BlockIdFlag),
		})
	}
	return ci
}
