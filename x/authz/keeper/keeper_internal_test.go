package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// internalRouter wraps a real router and marks chosen type URLs as internal.
type internalRouter struct {
	baseapp.MessageRouter
	internal     map[string]bool
	handlerCalls int
}

func (r *internalRouter) Handler(msg sdk.Msg) baseapp.MsgServiceHandler {
	r.handlerCalls++
	return r.MessageRouter.Handler(msg)
}

func (r *internalRouter) IsInternal(typeURL string) bool { return r.internal[typeURL] }

func (s *TestSuite) TestDispatchActionsRejectsInternalMsg() {
	msg := banktypes.NewMsgSend(s.addrs[0], s.addrs[1], coins10)
	typeURL := sdk.MsgTypeURL(msg)

	newKeeper := func(r *internalRouter) authzkeeper.Keeper {
		return authzkeeper.NewKeeper(s.storeService, s.encCfg.Codec, r, s.accountKeeper)
	}

	s.Run("internal msg rejected even when granter == grantee", func() {
		r := &internalRouter{MessageRouter: s.baseApp.MsgServiceRouter(), internal: map[string]bool{typeURL: true}}
		_, err := newKeeper(r).DispatchActions(s.ctx, s.addrs[0], []sdk.Msg{msg})
		s.Require().ErrorIs(err, sdkerrors.ErrUnauthorized)
		s.Require().ErrorContains(err, "internal-only")
		s.Require().Zero(r.handlerCalls, "handler must not be reached")
	})

	s.Run("control: external msg without grant fails for a different reason", func() {
		r := &internalRouter{MessageRouter: s.baseApp.MsgServiceRouter(), internal: map[string]bool{}}
		_, err := newKeeper(r).DispatchActions(s.ctx, s.addrs[1], []sdk.Msg{msg})
		s.Require().ErrorIs(err, authz.ErrNoAuthorizationFound)
		s.Require().NotContains(err.Error(), "internal-only")
	})
}
