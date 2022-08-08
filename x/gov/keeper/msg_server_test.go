package keeper_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/regen-network/gocuke"
	"github.com/stretchr/testify/require"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/gov/keeper"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

type msgServerSuite struct {
	*baseSuite

	msgServer v1.MsgServer
	err       error
}

func TestMsgServer(t *testing.T) {
	gocuke.NewRunner(t, &msgServerSuite{}).Path("./features/msg_submit_proposal.feature").Run()
}

func (s *msgServerSuite) Before(t gocuke.TestingT) {
	s.baseSuite = setupBase(t)
	s.msgServer = keeper.NewMsgServerImpl(s.govKeeper)
	s.expectCalls()
}

func (s *msgServerSuite) AMindepositParamSetToAndMininitialdepositrationSetTo(d string, r string) {
	p := v1.DefaultParams()
	coins, err := sdk.ParseCoinsNormalized(d)
	require.NoError(s.t, err)
	_, err = sdk.NewDecFromStr(r)
	require.NoError(s.t, err)

	p.MinDeposit = coins
	p.MinInitialDepositRatio = r

	err = s.govKeeper.SetParams(s.ctx, p)
	require.NoError(s.t, err)
}

func (s *msgServerSuite) AliceSubmitsAProposalWithMsg(a gocuke.DocString) {
	var msg sdk.Msg
	err := s.cdc.UnmarshalInterfaceJSON([]byte(a.Content), &msg)
	require.NoError(s.t, err)

	any, err := codectypes.NewAnyWithValue(msg)
	require.NoError(s.t, err)
	_, s.err = s.msgServer.SubmitProposal(sdk.WrapSDKContext(s.ctx), &v1.MsgSubmitProposal{
		Messages:       []*codectypes.Any{any},
		InitialDeposit: sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(10))),
		Proposer:       s.alice.String(),
	})
}

func (s *msgServerSuite) AliceSubmitsAProposalWithDeposit(d string) {
	coins, err := sdk.ParseCoinsNormalized(d)
	require.NoError(s.t, err)

	_, s.err = s.msgServer.SubmitProposal(sdk.WrapSDKContext(s.ctx), &v1.MsgSubmitProposal{
		Messages:       []*codectypes.Any{},
		InitialDeposit: coins,
		Proposer:       s.alice.String(),
	})
}

func (s *msgServerSuite) ExpectTheError(errStr string) {
	require.Error(s.t, s.err)
	require.Contains(s.t, s.err.Error(), errStr)
}

func (s *msgServerSuite) ExpectNoError() {
	require.NoError(s.t, s.err)
}

func (s *msgServerSuite) expectCalls() {
	s.authKeeper.EXPECT().GetModuleAccount(s.ctx, types.ModuleName).Return(authtypes.NewEmptyModuleAccount(types.ModuleName)).AnyTimes()
	s.bankKeeper.EXPECT().SendCoinsFromAccountToModule(s.ctx, s.alice, types.ModuleName, gomock.Any()).Return(nil).AnyTimes()
}
