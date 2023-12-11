package simapp_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	cmtcfg "github.com/cometbft/cometbft/config"
	"github.com/cosmos/go-bip39"
	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/simapp"
	authsim "cosmossdk.io/x/auth/simulation"
	banksim "cosmossdk.io/x/bank/simulation"
	govsim "cosmossdk.io/x/gov/simulation"

	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	xsim "github.com/cosmos/cosmos-sdk/x/simulation"
)

type IntegrationTestSuite struct {
	suite.Suite

	network  network.NetworkI
	cfg      network.Config
	v0app    *simapp.SimApp
	rand     *rand.Rand
	accounts []simtypes.Account
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	s.rand = rand.New(rand.NewSource(42))
	params := xsim.RandomParams(s.rand)
	s.accounts = simtypes.RandomAccounts(s.rand, params.NumKeys())
	// TODO: improve this hack to get at the bank keeper, or replace
	// it with RPC queries.
	s.cfg = network.DefaultConfig(func() network.TestFixture {
		fixture := simapp.NewTestNetworkFixture()
		appCtr := fixture.AppConstructor
		fixture.AppConstructor = func(val network.ValidatorI) servertypes.Application {
			app := appCtr(val)
			s.v0app = app.(*simapp.SimApp)
			return app
		}
		return fixture
	})
	s.cfg.NumValidators = s.rand.Intn(10)
	s.cfg.ChainID = fmt.Sprintf("chainid-simapp-%x", s.rand.Uint64())
	s.cfg.GenesisTime = time.Unix(0, 0)
	s.cfg.FuzzConnConfig = cmtcfg.DefaultFuzzConnConfig()
	s.cfg.FuzzConnConfig.ProbDropConn = .1
	for i := 0; i < s.cfg.NumValidators; i++ {
		entropy := make([]byte, 256/8)
		n, _ := s.rand.Read(entropy)
		mnemonic, err := bip39.NewMnemonic(entropy[:n])
		if err != nil {
			s.T().Fatal(err)
		}
		s.cfg.Mnemonics = append(s.cfg.Mnemonics, mnemonic)
	}
	initialStake := sdkmath.NewInt(s.rand.Int63n(1e12))
	const numBonded = 3
	banksim.RandomizedGenState(s.rand, s.cfg.GenesisState, s.cfg.Codec, s.cfg.BondDenom, s.accounts)
	authsim.RandomizedGenState(s.rand, s.cfg.GenesisState, s.cfg.Codec, s.cfg.BondDenom, s.accounts, initialStake, numBonded, s.cfg.GenesisTime)
	govsim.RandomizedGenState(s.rand, s.cfg.GenesisState, s.cfg.Codec, s.cfg.BondDenom)
	var err error
	s.network, err = network.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err)

	h, err := s.network.WaitForHeight(1)
	s.Require().NoError(err, "stalled at height %d", h)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) TestNetwork_Liveness() {
	h, err := s.network.WaitForHeightWithTimeout(10, time.Minute)
	s.Require().NoError(err, "expected to reach 10 blocks; got %d", h)
}

func (s *IntegrationTestSuite) TestSimulation() {
	v0 := s.network.GetValidators()[0]
	v0app := s.v0app
	clientCtx := v0.GetClientCtx()
	ctx := v0app.NewContext(true)

	voteGen := govsim.NewVoteGenerator()
	var futureVotes []govsim.Vote
	txEnc := clientCtx.TxConfig.TxEncoder()
	queryClient := cmtservice.NewServiceClient(clientCtx)
	generate := func(r *rand.Rand) sdk.Tx {
		bankSend := func(to, from simtypes.Account) sdk.Tx {
			return banksim.GenerateMsgSend(s.rand, ctx, clientCtx, v0.GetMoniker(), to, from, v0app.BankKeeper, v0app.AuthKeeper)
		}
		generators := []func() sdk.Tx{
			func() sdk.Tx {
				from, _ := simtypes.RandomAcc(s.rand, s.accounts)
				to, _ := simtypes.RandomAcc(s.rand, s.accounts)
				// disallow sending money to yourself.
				for from.PubKey.Equals(to.PubKey) {
					to, _ = simtypes.RandomAcc(s.rand, s.accounts)
				}
				return bankSend(to, from)
			},
			func() sdk.Tx {
				macc := v0app.AuthKeeper.GetModuleAccount(ctx, banksim.DistributionModuleName)
				to := simtypes.Account{
					PubKey:  macc.GetPubKey(),
					Address: macc.GetAddress(),
				}
				from, _ := simtypes.RandomAcc(s.rand, s.accounts)
				return bankSend(to, from)
			},
			func() sdk.Tx {
				const numModuleAccs = 2
				return banksim.GenerateMsgMultiSendToModuleAccount(s.rand, ctx, clientCtx, v0.GetMoniker(), s.accounts, v0app.BankKeeper, v0app.AuthKeeper, numModuleAccs)
			},
			func() sdk.Tx {
				proposal := authsim.GenerateMsgUpdateParams(s.rand)
				tx := govsim.GenerateMsgSubmitProposal(s.rand, ctx, clientCtx.TxConfig, s.accounts, v0app.AuthKeeper, v0app.BankKeeper, v0app.GovKeeper, []sdk.Msg{proposal})
				votes := voteGen.GenerateVotes(s.rand, ctx, clientCtx.TxConfig, s.accounts, v0app.AuthKeeper, v0app.BankKeeper, v0app.GovKeeper)
				futureVotes = append(futureVotes, votes...)
				return tx
			},
			func() sdk.Tx {
				return govsim.GenerateMsgDeposit(s.rand, ctx, clientCtx.TxConfig, s.accounts, v0app.AuthKeeper, v0app.BankKeeper, v0app.GovKeeper)
			},
			func() sdk.Tx {
				return banksim.GenerateMsgMultiSend(s.rand, ctx, clientCtx, v0.GetMoniker(), s.accounts, v0app.BankKeeper, v0app.AuthKeeper)
			},
			func() sdk.Tx {
				return govsim.GenerateMsgVoteWeighted(s.rand, ctx, clientCtx.TxConfig, s.accounts, v0app.AuthKeeper, v0app.BankKeeper, v0app.GovKeeper)
			},
			func() sdk.Tx {
				return govsim.GenerateMsgCancelProposal(s.rand, ctx, clientCtx.TxConfig, s.accounts, v0app.AuthKeeper, v0app.BankKeeper, v0app.GovKeeper)
			},
		}
		idx := s.rand.Intn(len(generators))
		return generators[idx]()
	}
	for i := 0; i < 100; i++ {
		tx := generate(s.rand)
		txBytes, err := txEnc(tx)
		s.Require().NoError(err)
		_, err = clientCtx.BroadcastTxAsync(txBytes)
		s.Require().NoError(err)
		res, err := queryClient.GetLatestBlock(ctx, &cmtservice.GetLatestBlockRequest{})
		s.Require().NoError(err)
		latestTime := res.SdkBlock.Header.Time
		for j := len(futureVotes) - 1; j >= 0; j-- {
			v := futureVotes[j]
			if v.BlockTime.Before(latestTime) {
				continue
			}
			futureVotes = append(futureVotes[:j], futureVotes[j+1:]...)
			txBytes, err := txEnc(v.Vote)
			s.Require().NoError(err)
			_, err = clientCtx.BroadcastTxAsync(txBytes)
			s.Require().NoError(err)
		}
		if i%10 == 0 {
			err := s.network.WaitForNextBlock()
			s.Require().NoError(err)
		}
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
