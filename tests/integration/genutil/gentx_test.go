package genutil_test

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"
	"time"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	_ "github.com/cosmos/cosmos-sdk/x/auth"
	_ "github.com/cosmos/cosmos-sdk/x/auth/tx/config"
	_ "github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	_ "github.com/cosmos/cosmos-sdk/x/consensus"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
	_ "github.com/cosmos/cosmos-sdk/x/params"
	_ "github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var (
	priv1 = secp256k1.GenPrivKey()
	priv2 = secp256k1.GenPrivKey()
	pk1   = priv1.PubKey()
	pk2   = priv2.PubKey()
	addr1 = sdk.AccAddress(pk1.Address())
	addr2 = sdk.AccAddress(pk2.Address())
	desc  = stakingtypes.NewDescription("testname", "", "", "", "")
	comm  = stakingtypes.CommissionRates{}
)

// GenTxTestSuite is a test suite to be used with gentx tests.
type GenTxTestSuite struct {
	suite.Suite

	ctx sdk.Context

	encodingConfig moduletestutil.TestEncodingConfig
	msg1, msg2     *stakingtypes.MsgCreateValidator
	accountKeeper  authkeeper.AccountKeeper
	bankKeeper     bankkeeper.Keeper
	stakingKeeper  *stakingkeeper.Keeper
	baseApp        *baseapp.BaseApp
}

func (suite *GenTxTestSuite) SetupTest() {
	encCfg := moduletestutil.TestEncodingConfig{}

	app, err := simtestutil.SetupWithConfiguration(
		configurator.NewAppConfig(
			configurator.BankModule(),
			configurator.TxModule(),
			configurator.StakingModule(),
			configurator.ParamsModule(),
			configurator.ConsensusModule(),
			configurator.AuthModule()),
		simtestutil.DefaultStartUpConfig(),
		&encCfg.InterfaceRegistry, &encCfg.Codec, &encCfg.TxConfig, &encCfg.Amino,
		&suite.accountKeeper, &suite.bankKeeper, &suite.stakingKeeper)
	suite.Require().NoError(err)

	suite.ctx = app.BaseApp.NewContext(false, tmproto.Header{})
	suite.encodingConfig = encCfg
	suite.baseApp = app.BaseApp

	amount := sdk.NewInt64Coin(sdk.DefaultBondDenom, 50)
	suite.msg1, err = stakingtypes.NewMsgCreateValidator(
		sdk.ValAddress(pk1.Address()), pk1, amount, desc, comm)
	suite.NoError(err)
	suite.msg2, err = stakingtypes.NewMsgCreateValidator(
		sdk.ValAddress(pk2.Address()), pk1, amount, desc, comm)
	suite.NoError(err)
}

func (suite *GenTxTestSuite) setAccountBalance(addr sdk.AccAddress, amount int64) json.RawMessage {
	acc := suite.accountKeeper.NewAccountWithAddress(suite.ctx, addr)
	suite.accountKeeper.SetAccount(suite.ctx, acc)

	err := testutil.FundAccount(suite.bankKeeper, suite.ctx, addr, sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, amount)})
	suite.Require().NoError(err)

	bankGenesisState := suite.bankKeeper.ExportGenesis(suite.ctx)
	bankGenesis, err := suite.encodingConfig.Amino.MarshalJSON(bankGenesisState) // TODO switch this to use Marshaler
	suite.Require().NoError(err)

	return bankGenesis
}

func (suite *GenTxTestSuite) TestSetGenTxsInAppGenesisState() {
	var (
		txBuilder = suite.encodingConfig.TxConfig.NewTxBuilder()
		genTxs    []sdk.Tx
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"one genesis transaction",
			func() {
				err := txBuilder.SetMsgs(suite.msg1)
				suite.Require().NoError(err)
				tx := txBuilder.GetTx()
				genTxs = []sdk.Tx{tx}
			},
			true,
		},
		{
			"two genesis transactions",
			func() {
				err := txBuilder.SetMsgs(suite.msg1, suite.msg2)
				suite.Require().NoError(err)
				tx := txBuilder.GetTx()
				genTxs = []sdk.Tx{tx}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest()
			cdc := suite.encodingConfig.Codec
			txJSONEncoder := suite.encodingConfig.TxConfig.TxJSONEncoder()

			tc.malleate()
			appGenesisState, err := genutil.SetGenTxsInAppGenesisState(cdc, txJSONEncoder, make(map[string]json.RawMessage), genTxs)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(appGenesisState[types.ModuleName])

				var genesisState types.GenesisState
				err := cdc.UnmarshalJSON(appGenesisState[types.ModuleName], &genesisState)
				suite.Require().NoError(err)
				suite.Require().NotNil(genesisState.GenTxs)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *GenTxTestSuite) TestValidateAccountInGenesis() {
	var (
		appGenesisState = make(map[string]json.RawMessage)
		coins           sdk.Coins
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"no accounts",
			func() {
				coins = sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)}
			},
			false,
		},
		{
			"account without balance in the genesis state",
			func() {
				coins = sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)}
				appGenesisState[banktypes.ModuleName] = suite.setAccountBalance(addr2, 50)
			},
			false,
		},
		{
			"account without enough funds of default bond denom",
			func() {
				coins = sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 50)}
				appGenesisState[banktypes.ModuleName] = suite.setAccountBalance(addr1, 25)
			},
			false,
		},
		{
			"account with enough funds of default bond denom",
			func() {
				coins = sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 10)}
				appGenesisState[banktypes.ModuleName] = suite.setAccountBalance(addr1, 25)
			},
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest()
			cdc := suite.encodingConfig.Codec

			suite.stakingKeeper.SetParams(suite.ctx, stakingtypes.DefaultParams())
			stakingGenesisState := suite.stakingKeeper.ExportGenesis(suite.ctx)
			suite.Require().Equal(stakingGenesisState.Params, stakingtypes.DefaultParams())
			stakingGenesis, err := cdc.MarshalJSON(stakingGenesisState) // TODO switch this to use Marshaler
			suite.Require().NoError(err)
			appGenesisState[stakingtypes.ModuleName] = stakingGenesis

			tc.malleate()
			err = genutil.ValidateAccountInGenesis(
				appGenesisState, banktypes.GenesisBalancesIterator{},
				addr1, coins, cdc,
			)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *GenTxTestSuite) TestDeliverGenTxs() {
	var (
		genTxs    []json.RawMessage
		txBuilder = suite.encodingConfig.TxConfig.NewTxBuilder()
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"no signature supplied",
			func() {
				err := txBuilder.SetMsgs(suite.msg1)
				suite.Require().NoError(err)

				genTxs = make([]json.RawMessage, 1)
				tx, err := suite.encodingConfig.TxConfig.TxJSONEncoder()(txBuilder.GetTx())
				suite.Require().NoError(err)
				genTxs[0] = tx
			},
			false,
		},
		{
			"success",
			func() {
				_ = suite.setAccountBalance(addr1, 50)
				_ = suite.setAccountBalance(addr2, 1)

				r := rand.New(rand.NewSource(time.Now().UnixNano()))
				msg := banktypes.NewMsgSend(addr1, addr2, sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 1)})
				tx, err := simtestutil.GenSignedMockTx(
					r,
					suite.encodingConfig.TxConfig,
					[]sdk.Msg{msg},
					sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 10)},
					simtestutil.DefaultGenTxGas,
					suite.ctx.ChainID(),
					[]uint64{7},
					[]uint64{0},
					priv1,
				)
				suite.Require().NoError(err)

				genTxs = make([]json.RawMessage, 1)
				genTx, err := suite.encodingConfig.TxConfig.TxJSONEncoder()(tx)
				suite.Require().NoError(err)
				genTxs[0] = genTx
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest()

			tc.malleate()

			if tc.expPass {
				suite.Require().NotPanics(func() {
					genutil.DeliverGenTxs(
						suite.ctx, genTxs, suite.stakingKeeper, suite.baseApp.DeliverTx,
						suite.encodingConfig.TxConfig,
					)
				})
			} else {
				_, err := genutil.DeliverGenTxs(
					suite.ctx, genTxs, suite.stakingKeeper, suite.baseApp.DeliverTx,
					suite.encodingConfig.TxConfig,
				)

				suite.Require().Error(err)
			}
		})
	}
}

func TestGenTxTestSuite(t *testing.T) {
	suite.Run(t, new(GenTxTestSuite))
}
