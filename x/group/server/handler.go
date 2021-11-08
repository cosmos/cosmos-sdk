package server

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	servermodule "github.com/cosmos/cosmos-sdk/types/module/server"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/cosmos/cosmos-sdk/x/group/exported"
)

// NewHandler creates an sdk.Handler for all the group type messages.
// This is needed for supporting amino-json signing.
func NewHandler(configurator servermodule.Configurator, accountKeeper exported.AccountKeeper, bankKeeper exported.BankKeeper) sdk.Handler {
	impl := newServer(configurator.ModuleKey(), accountKeeper, bankKeeper, configurator.Marshaler())

	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case *group.MsgCreateGroup:
			res, err := impl.CreateGroup(ctx.Context(), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *group.MsgUpdateGroupMembers:
			res, err := impl.UpdateGroupMembers(ctx.Context(), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *group.MsgUpdateGroupAdmin:
			res, err := impl.UpdateGroupAdmin(ctx.Context(), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *group.MsgUpdateGroupMetadata:
			res, err := impl.UpdateGroupMetadata(ctx.Context(), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *group.MsgCreateGroupAccount:
			res, err := impl.CreateGroupAccount(ctx.Context(), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *group.MsgUpdateGroupAccountAdmin:
			res, err := impl.UpdateGroupAccountAdmin(ctx.Context(), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *group.MsgUpdateGroupAccountDecisionPolicy:
			res, err := impl.UpdateGroupAccountDecisionPolicy(ctx.Context(), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *group.MsgUpdateGroupAccountMetadata:
			res, err := impl.UpdateGroupAccountMetadata(ctx.Context(), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *group.MsgCreateProposal:
			res, err := impl.CreateProposal(ctx.Context(), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *group.MsgVote:
			res, err := impl.Vote(ctx.Context(), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *group.MsgExec:
			res, err := impl.Exec(ctx.Context(), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		default:
			return nil, errors.Wrapf(errors.ErrUnknownRequest, "unrecognized %s message type: %T", group.ModuleName, msg)
		}
	}
}
