package genutil_test

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"cosmossdk.io/math"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"
	"gotest.tools/v3/assert"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	_ "github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	_ "github.com/cosmos/cosmos-sdk/x/auth/tx/config"
	_ "github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	_ "github.com/cosmos/cosmos-sdk/x/consensus"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
	_ "github.com/cosmos/cosmos-sdk/x/params"
	_ "github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
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

type fixture struct {
	ctx sdk.Context

	encodingConfig moduletestutil.TestEncodingConfig
	msg1, msg2     *stakingtypes.MsgCreateValidator
	accountKeeper  authkeeper.AccountKeeper
	bankKeeper     bankkeeper.Keeper
	stakingKeeper  *stakingkeeper.Keeper
	baseApp        *baseapp.BaseApp
}

func initFixture(t assert.TestingT) *fixture {
	f := &fixture{}
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
		&f.accountKeeper, &f.bankKeeper, &f.stakingKeeper)
	assert.NilError(t, err)

	f.ctx = app.BaseApp.NewContext(false, cmtproto.Header{})
	f.encodingConfig = encCfg
	f.baseApp = app.BaseApp

	amount := sdk.NewInt64Coin(sdk.DefaultBondDenom, 50)
	one := math.OneInt()
	f.msg1, err = stakingtypes.NewMsgCreateValidator(
		sdk.ValAddress(pk1.Address()), pk1, amount, desc, comm, one)
	assert.NilError(t, err)
	f.msg2, err = stakingtypes.NewMsgCreateValidator(
		sdk.ValAddress(pk2.Address()), pk1, amount, desc, comm, one)
	assert.NilError(t, err)

	return f
}

func setAccountBalance(t *testing.T, f *fixture, addr sdk.AccAddress, amount int64) json.RawMessage {
	acc := f.accountKeeper.NewAccountWithAddress(f.ctx, addr)
	f.accountKeeper.SetAccount(f.ctx, acc)

	err := testutil.FundAccount(f.bankKeeper, f.ctx, addr, sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, amount)})
	assert.NilError(t, err)

	bankGenesisState := f.bankKeeper.ExportGenesis(f.ctx)
	bankGenesis, err := f.encodingConfig.Amino.MarshalJSON(bankGenesisState) // TODO switch this to use Marshaler
	assert.NilError(t, err)

	return bankGenesis
}

func TestSetGenTxsInAppGenesisState(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	var (
		txBuilder = f.encodingConfig.TxConfig.NewTxBuilder()
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
				err := txBuilder.SetMsgs(f.msg1)
				assert.NilError(t, err)
				tx := txBuilder.GetTx()
				genTxs = []sdk.Tx{tx}
			},
			true,
		},
		{
			"two genesis transactions",
			func() {
				err := txBuilder.SetMsgs(f.msg1, f.msg2)
				assert.NilError(t, err)
				tx := txBuilder.GetTx()
				genTxs = []sdk.Tx{tx}
			},
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.msg), func(t *testing.T) {
			initFixture(t)

			cdc := f.encodingConfig.Codec
			txJSONEncoder := f.encodingConfig.TxConfig.TxJSONEncoder()

			tc.malleate()
			appGenesisState, err := genutil.SetGenTxsInAppGenesisState(cdc, txJSONEncoder, make(map[string]json.RawMessage), genTxs)

			assert.NilError(t, err)
			assert.Assert(t, appGenesisState[types.ModuleName] != nil)

			var genesisState types.GenesisState
			err = cdc.UnmarshalJSON(appGenesisState[types.ModuleName], &genesisState)
			assert.NilError(t, err)
			assert.Assert(t, genesisState.GenTxs != nil)
		})
	}
}

func TestValidateAccountInGenesis(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	var (
		appGenesisState = make(map[string]json.RawMessage)
		coins           sdk.Coins
	)

	testCases := []struct {
		msg       string
		malleate  func()
		expPass   bool
		expErrMsg string
	}{
		{
			"account without balance in the genesis state",
			func() {
				coins = sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)}
			},
			false,
			fmt.Sprintf("account %s does not have a balance in the genesis state", addr1),
		},
		{
			"account without enough funds of default bond denom",
			func() {
				coins = sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 50)}
				appGenesisState[banktypes.ModuleName] = setAccountBalance(t, f, addr1, 25)
			},
			false,
			fmt.Sprintf("account %s has a balance in genesis, but it only has 25stake available to stake, not 50stake", addr1),
		},
		{
			"account with enough funds of default bond denom",
			func() {
				coins = sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 10)}
				appGenesisState[banktypes.ModuleName] = setAccountBalance(t, f, addr1, 25)
			},
			true,
			"",
		},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.msg), func(t *testing.T) {
			initFixture(t)

			cdc := f.encodingConfig.Codec

			f.stakingKeeper.SetParams(f.ctx, stakingtypes.DefaultParams())
			stakingGenesisState := f.stakingKeeper.ExportGenesis(f.ctx)
			assert.DeepEqual(t, stakingGenesisState.Params, stakingtypes.DefaultParams())
			stakingGenesis, err := cdc.MarshalJSON(stakingGenesisState) // TODO switch this to use Marshaler
			assert.NilError(t, err)
			appGenesisState[stakingtypes.ModuleName] = stakingGenesis

			tc.malleate()
			err = genutil.ValidateAccountInGenesis(
				appGenesisState, banktypes.GenesisBalancesIterator{},
				addr1, coins, cdc,
			)

			if tc.expPass {
				assert.NilError(t, err)
			} else {
				assert.ErrorContains(t, err, tc.expErrMsg)
			}
		})
	}
}

func TestDeliverGenTxs(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	var (
		genTxs    []json.RawMessage
		txBuilder = f.encodingConfig.TxConfig.NewTxBuilder()
	)

	testCases := []struct {
		msg       string
		malleate  func()
		expPass   bool
		expErrMsg string
	}{
		{
			"no signature supplied",
			func() {
				err := txBuilder.SetMsgs(f.msg1)
				assert.NilError(t, err)

				genTxs = make([]json.RawMessage, 1)
				tx, err := f.encodingConfig.TxConfig.TxJSONEncoder()(txBuilder.GetTx())
				assert.NilError(t, err)
				genTxs[0] = tx
			},
			false,
			"no signatures supplied",
		},
		{
			"success",
			func() {
				_ = setAccountBalance(t, f, addr1, 50)
				_ = setAccountBalance(t, f, addr2, 1)

				r := rand.New(rand.NewSource(time.Now().UnixNano()))
				msg := banktypes.NewMsgSend(addr1, addr2, sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 1)})
				tx, err := simtestutil.GenSignedMockTx(
					r,
					f.encodingConfig.TxConfig,
					[]sdk.Msg{msg},
					sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 10)},
					simtestutil.DefaultGenTxGas,
					f.ctx.ChainID(),
					[]uint64{7},
					[]uint64{0},
					priv1,
				)
				assert.NilError(t, err)

				genTxs = make([]json.RawMessage, 1)
				genTx, err := f.encodingConfig.TxConfig.TxJSONEncoder()(tx)
				assert.NilError(t, err)
				genTxs[0] = genTx
			},
			true,
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.msg), func(t *testing.T) {
			initFixture(t)

			tc.malleate()

			if tc.expPass {
				require.NotPanics(t, func() {
					genutil.DeliverGenTxs(
						f.ctx, genTxs, f.stakingKeeper, f.baseApp.DeliverTx,
						f.encodingConfig.TxConfig,
					)
				})
			} else {
				_, err := genutil.DeliverGenTxs(
					f.ctx, genTxs, f.stakingKeeper, f.baseApp.DeliverTx,
					f.encodingConfig.TxConfig,
				)

				assert.ErrorContains(t, err, tc.expErrMsg)
			}
		})
	}
}
