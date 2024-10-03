package multisig

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/math"
	"cosmossdk.io/simapp"
	v1 "cosmossdk.io/x/accounts/defaults/multisig/v1"
	"cosmossdk.io/x/bank/testutil"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type E2ETestSuite struct {
	suite.Suite

	app         *simapp.SimApp
	members     []sdk.AccAddress
	membersAddr []string
}

func NewE2ETestSuite() *E2ETestSuite {
	return &E2ETestSuite{}
}

func (s *E2ETestSuite) SetupSuite() {
	s.app = setupApp(s.T())

	s.members = []sdk.AccAddress{}
	for i := 0; i < 10; i++ {
		addr := secp256k1.GenPrivKey().PubKey().Address()
		addrStr, err := s.app.AuthKeeper.AddressCodec().BytesToString(addr)
		require.NoError(s.T(), err)
		s.membersAddr = append(s.membersAddr, addrStr)
		s.members = append(s.members, sdk.AccAddress(addr))
	}
}

func (s *E2ETestSuite) TearDownSuite() {}

func setupApp(t *testing.T) *simapp.SimApp {
	t.Helper()
	app := simapp.Setup(t, false)
	return app
}

func (s *E2ETestSuite) executeTx(ctx context.Context, msg sdk.Msg, accAddr, sender []byte) error {
	_, err := s.app.AccountsKeeper.Execute(ctx, accAddr, sender, msg, nil)
	return err
}

func (s *E2ETestSuite) queryAcc(ctx context.Context, req sdk.Msg, accAddr []byte) (transaction.Msg, error) {
	resp, err := s.app.AccountsKeeper.Query(ctx, accAddr, req)
	return resp, err
}

func (s *E2ETestSuite) fundAccount(ctx context.Context, addr sdk.AccAddress, amt sdk.Coins) {
	require.NoError(s.T(), testutil.FundAccount(ctx, s.app.BankKeeper, addr, amt))
}

// initAccount initializes a multisig account with the given members and powers
// and returns the account address
func (s *E2ETestSuite) initAccount(ctx context.Context, sender []byte, membersPowers map[string]uint64) ([]byte, string) {
	s.fundAccount(ctx, sender, sdk.Coins{sdk.NewCoin("stake", math.NewInt(1000000))})

	members := []*v1.Member{}
	for addrStr, power := range membersPowers {
		members = append(members, &v1.Member{Address: addrStr, Weight: power})
	}

	_, accountAddr, err := s.app.AccountsKeeper.Init(ctx, "multisig", sender,
		&v1.MsgInit{
			Members: members,
			Config: &v1.Config{
				Threshold:      100,
				Quorum:         100,
				VotingPeriod:   120,
				Revote:         false,
				EarlyExecution: true,
			},
		}, sdk.Coins{sdk.NewCoin("stake", math.NewInt(1000))})
	s.NoError(err)

	accountAddrStr, err := s.app.AuthKeeper.AddressCodec().BytesToString(accountAddr)
	s.NoError(err)

	return accountAddr, accountAddrStr
}

// createProposal
func (s *E2ETestSuite) createProposal(ctx context.Context, accAddr, sender []byte, msgs ...*codectypes.Any) {
	propReq := &v1.MsgCreateProposal{
		Proposal: &v1.Proposal{
			Title:    "test",
			Summary:  "test",
			Messages: msgs,
		},
	}
	err := s.executeTx(ctx, propReq, accAddr, sender)
	s.NoError(err)
}

func (s *E2ETestSuite) executeProposal(ctx context.Context, accAddr, sender []byte, proposalID uint64) error {
	execReq := &v1.MsgExecuteProposal{
		ProposalId: proposalID,
	}
	return s.executeTx(ctx, execReq, accAddr, sender)
}
