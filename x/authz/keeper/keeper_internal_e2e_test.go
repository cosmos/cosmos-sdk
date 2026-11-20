package keeper_test

import (
	"context"

	baseapptestutil "github.com/cosmos/cosmos-sdk/baseapp/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
)

type countingInternalServer struct{ calls *int }

func (c countingInternalServer) DoInternal(_ context.Context, _ *baseapptestutil.MsgInternal) (*baseapptestutil.MsgInternalResponse, error) {
	*c.calls++
	return &baseapptestutil.MsgInternalResponse{}, nil
}

func (s *TestSuite) TestExecRefusesInternalMsgEndToEnd() {
	calls := 0
	baseapptestutil.RegisterInterfaces(s.encCfg.InterfaceRegistry)
	router := s.baseApp.MsgServiceRouter()
	baseapptestutil.RegisterInternalServer(router, countingInternalServer{calls: &calls})
	authz.RegisterMsgServer(router, s.authzKeeper)

	msg := &baseapptestutil.MsgInternal{Signer: s.addrs[0].String()}

	s.Run("MsgExec", func() {
		exec := authz.NewMsgExec(s.addrs[0], []sdk.Msg{msg})
		_, err := s.authzKeeper.Exec(s.ctx, &exec)
		s.Require().ErrorIs(err, sdkerrors.ErrUnauthorized)
		s.Require().ErrorContains(err, "internal-only")
		s.Require().Zero(calls)
	})

	s.Run("nested MsgExec", func() {
		inner := authz.NewMsgExec(s.addrs[0], []sdk.Msg{msg})
		outer := authz.NewMsgExec(s.addrs[0], []sdk.Msg{&inner})
		_, err := s.authzKeeper.Exec(s.ctx, &outer)
		s.Require().ErrorIs(err, sdkerrors.ErrUnauthorized)
		s.Require().ErrorContains(err, "internal-only")
		s.Require().Zero(calls)
	})
}
