package simulation_test

import (
	"math/rand"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/depinject"
	"cosmossdk.io/log/v2"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/mldsa65"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	_ "github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/simulation"
	_ "github.com/cosmos/cosmos-sdk/x/auth/tx/config"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	_ "github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	_ "github.com/cosmos/cosmos-sdk/x/consensus"
	simulationx "github.com/cosmos/cosmos-sdk/x/simulation"
	_ "github.com/cosmos/cosmos-sdk/x/staking"
)

type SimTestSuite struct {
	suite.Suite

	ctx           sdk.Context
	accountKeeper authkeeper.AccountKeeper
	bankKeeper    bankkeeper.Keeper
	cdc           codec.Codec
	txConfig      client.TxConfig
	app           *runtime.App
}

func (s *SimTestSuite) SetupTest() {
	var err error
	s.app, err = simtestutil.Setup(
		depinject.Configs(
			configurator.NewAppConfig(
				configurator.AuthModule(),
				configurator.BankModule(),
				configurator.StakingModule(),
				configurator.ConsensusModule(),
				configurator.TxModule(),
			),
			depinject.Supply(log.NewNopLogger()),
		), &s.accountKeeper, &s.bankKeeper, &s.cdc, &s.txConfig)
	s.Require().NoError(err)

	s.ctx = s.app.NewContext(false)
}

func (s *SimTestSuite) setPubKeyChangeEnabled(enabled bool) {
	params := s.accountKeeper.GetParams(s.ctx)
	params.PubKeyChangeEnabled = enabled
	s.Require().NoError(s.accountKeeper.Params.Set(s.ctx, params))
}

func (s *SimTestSuite) getTestingAccounts(r *rand.Rand, n int) []simtypes.Account {
	accounts := simtypes.RandomAccounts(r, n)

	initAmt := sdk.TokensFromConsensusPower(200, sdk.DefaultPowerReduction)
	initCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initAmt))

	for _, account := range accounts {
		acc := s.accountKeeper.NewAccountWithAddress(s.ctx, account.Address)
		s.accountKeeper.SetAccount(s.ctx, acc)
		s.Require().NoError(banktestutil.FundAccount(s.ctx, s.bankKeeper, account.Address, initCoins))
	}

	return accounts
}

func (s *SimTestSuite) finalizeBlock() {
	_, err := s.app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: s.app.LastBlockHeight() + 1,
		Hash:   s.app.LastCommitID().Hash,
	})
	s.Require().NoError(err)
}

// TestWeightedOperations tests the weights of the operations.
func (s *SimTestSuite) TestWeightedOperations() {
	weightedOps := simulation.WeightedOperations(make(simtypes.AppParams), s.txConfig, s.accountKeeper)
	s.Require().Len(weightedOps, 1)
	s.Require().Equal(simulation.DefaultWeightMsgChangePubKey, weightedOps[0].Weight())
}

// TestSimulateMsgChangePubKey rotates a simulation account to a new key of
// each supported type and checks that the account can still sign txs with the
// private key the simulation now holds for it.
func (s *SimTestSuite) TestSimulateMsgChangePubKey() {
	genSecp256k1 := func(r *rand.Rand) (cryptotypes.PrivKey, error) {
		seed := make([]byte, 15)
		_, _ = r.Read(seed)
		return secp256k1.GenPrivKeyFromSecret(seed), nil
	}
	genMlDsa65 := func(r *rand.Rand) (cryptotypes.PrivKey, error) {
		seed := make([]byte, 32)
		_, _ = r.Read(seed)
		priv, err := mldsa65.GenPrivKeyFromSeed(seed)
		return &priv, err
	}

	for name, tc := range map[string]struct {
		genKey  func(*rand.Rand) (cryptotypes.PrivKey, error)
		keyType cryptotypes.PubKey
	}{
		"secp256k1": {genSecp256k1, &secp256k1.PubKey{}},
		"mldsa65":   {genMlDsa65, &mldsa65.PubKey{}},
	} {
		s.Run(name, func() {
			s.SetupTest()
			r := rand.New(rand.NewSource(1))
			accounts := s.getTestingAccounts(r, 3)
			before := append([]simtypes.Account(nil), accounts...)
			s.setPubKeyChangeEnabled(true)
			s.finalizeBlock()

			op := simulation.SimulateMsgChangePubKeyWithKeyGen(s.txConfig, s.accountKeeper, tc.genKey)
			opMsg, futureOps, err := op(r, s.app.BaseApp, s.ctx, accounts, "")
			s.Require().NoError(err)
			s.Require().True(opMsg.OK, opMsg.Comment)
			s.Require().Equal(sdk.MsgTypeURL(&types.MsgChangePubKey{}), opMsg.Name)
			s.Require().Len(futureOps, 0)

			// Exactly one account was rotated, and the simulation holds the
			// private key for its new on-chain pubkey.
			rotated := -1
			for i := range accounts {
				s.Require().True(accounts[i].Address.Equals(before[i].Address))
				if !accounts[i].PubKey.Equals(before[i].PubKey) {
					s.Require().Equal(-1, rotated, "more than one account rotated")
					rotated = i
				}
			}
			s.Require().NotEqual(-1, rotated, "no account rotated")
			simAcc := accounts[rotated]
			s.Require().IsType(tc.keyType, simAcc.PubKey)
			s.Require().True(simAcc.PrivKey.PubKey().Equals(simAcc.PubKey))
			onChain := s.accountKeeper.GetAccount(s.ctx, simAcc.Address)
			s.Require().True(simAcc.PubKey.Equals(onChain.GetPubKey()))

			// A later tx signed with the synced key is accepted.
			to := accounts[(rotated+1)%len(accounts)]
			sendMsg := banktypes.NewMsgSend(simAcc.Address, to.Address, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1)))
			opMsg, _, err = simulationx.GenAndDeliverTx(simulationx.OperationInput{
				R:             r,
				App:           s.app.BaseApp,
				TxGen:         s.txConfig,
				Msg:           sendMsg,
				Context:       s.ctx,
				SimAccount:    simAcc,
				AccountKeeper: s.accountKeeper,
				ModuleName:    banktypes.ModuleName,
			}, nil)
			s.Require().NoError(err)
			s.Require().True(opMsg.OK)
		})
	}
}

// TestSimulateMsgChangePubKeyDisabled checks that the operation is a no-op
// while account rekeying is disabled.
func (s *SimTestSuite) TestSimulateMsgChangePubKeyDisabled() {
	r := rand.New(rand.NewSource(1))
	accounts := s.getTestingAccounts(r, 3)
	before := append([]simtypes.Account(nil), accounts...)
	s.setPubKeyChangeEnabled(false)
	s.finalizeBlock()

	op := simulation.SimulateMsgChangePubKey(s.txConfig, s.accountKeeper)
	opMsg, futureOps, err := op(r, s.app.BaseApp, s.ctx, accounts, "")
	s.Require().NoError(err)
	s.Require().False(opMsg.OK)
	s.Require().Len(futureOps, 0)
	s.Require().Equal(before, accounts)
}

func TestSimTestSuite(t *testing.T) {
	suite.Run(t, new(SimTestSuite))
}
