package staking_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/staking/types"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
)

func bootstrapGenesisTest(t *testing.T, power int64, numAddrs int) (*simapp.SimApp, sdk.Context, []sdk.AccAddress) {
	_, app, ctx := getBaseSimappWithCustomKeeper()

	addrDels, _ := generateAddresses(app, ctx, numAddrs, 10000)

	amt := sdk.TokensFromConsensusPower(power)
	totalSupply := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), amt.MulRaw(int64(len(addrDels)))))

	notBondedPool := app.StakingKeeper.GetNotBondedPool(ctx)
	err := app.BankKeeper.SetBalances(ctx, notBondedPool.GetAddress(), totalSupply)
	require.NoError(t, err)

	app.AccountKeeper.SetModuleAccount(ctx, notBondedPool)
	app.BankKeeper.SetSupply(ctx, banktypes.NewSupply(totalSupply))

	return app, ctx, addrDels
}

func TestInitGenesis(t *testing.T) {
	app, ctx, addrs := bootstrapGenesisTest(t, 1000, 10)

	valTokens := sdk.TokensFromConsensusPower(1)

	params := app.StakingKeeper.GetParams(ctx)
	validators := make([]types.Validator, 2)
	var delegations []types.Delegation

	pk0, err := sdk.Bech32ifyPubKey(sdk.Bech32PubKeyTypeConsPub, PKs[0])
	require.NoError(t, err)

	pk1, err := sdk.Bech32ifyPubKey(sdk.Bech32PubKeyTypeConsPub, PKs[1])
	require.NoError(t, err)

	// initialize the validators
	validators[0].OperatorAddress = sdk.ValAddress(addrs[0])
	validators[0].ConsensusPubkey = pk0
	validators[0].Description = types.NewDescription("hoop", "", "", "", "")
	validators[0].Status = sdk.Bonded
	validators[0].Tokens = valTokens
	validators[0].DelegatorShares = valTokens.ToDec()
	validators[1].OperatorAddress = sdk.ValAddress(addrs[1])
	validators[1].ConsensusPubkey = pk1
	validators[1].Description = types.NewDescription("bloop", "", "", "", "")
	validators[1].Status = sdk.Bonded
	validators[1].Tokens = valTokens
	validators[1].DelegatorShares = valTokens.ToDec()

	genesisState := types.NewGenesisState(params, validators, delegations)
	vals := staking.InitGenesis(ctx, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, genesisState)

	actualGenesis := staking.ExportGenesis(ctx, app.StakingKeeper)
	require.Equal(t, genesisState.Params, actualGenesis.Params)
	require.Equal(t, genesisState.Delegations, actualGenesis.Delegations)
	require.EqualValues(t, app.StakingKeeper.GetAllValidators(ctx), actualGenesis.Validators)

	// Ensure validators have addresses.
	for _, val := range staking.WriteValidators(ctx, app.StakingKeeper) {
		require.NotEmpty(t, val.Address)
	}

	// now make sure the validators are bonded and intra-tx counters are correct
	resVal, found := app.StakingKeeper.GetValidator(ctx, sdk.ValAddress(addrs[0]))
	require.True(t, found)
	require.Equal(t, sdk.Bonded, resVal.Status)

	resVal, found = app.StakingKeeper.GetValidator(ctx, sdk.ValAddress(addrs[1]))
	require.True(t, found)
	require.Equal(t, sdk.Bonded, resVal.Status)

	abcivals := make([]abci.ValidatorUpdate, len(vals))
	for i, val := range validators {
		abcivals[i] = val.ABCIValidatorUpdate()
	}

	require.Equal(t, abcivals, vals)
}

func TestInitGenesisLargeValidatorSet(t *testing.T) {
	size := 200
	require.True(t, size > 100)

	app, ctx, addrs := bootstrapGenesisTest(t, 1000, 200)

	params := app.StakingKeeper.GetParams(ctx)
	delegations := []types.Delegation{}
	validators := make([]types.Validator, size)

	for i := range validators {
		validators[i] = types.NewValidator(sdk.ValAddress(addrs[i]),
			PKs[i], types.NewDescription(fmt.Sprintf("#%d", i), "", "", "", ""))

		validators[i].Status = sdk.Bonded

		tokens := sdk.TokensFromConsensusPower(1)
		if i < 100 {
			tokens = sdk.TokensFromConsensusPower(2)
		}
		validators[i].Tokens = tokens
		validators[i].DelegatorShares = tokens.ToDec()
	}

	genesisState := types.NewGenesisState(params, validators, delegations)
	vals := staking.InitGenesis(ctx, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, genesisState)

	abcivals := make([]abci.ValidatorUpdate, 100)
	for i, val := range validators[:100] {
		abcivals[i] = val.ABCIValidatorUpdate()
	}

	require.Equal(t, abcivals, vals)
}

func TestValidateGenesis(t *testing.T) {
	genValidators1 := make([]types.Validator, 1, 5)
	pk := ed25519.GenPrivKey().PubKey()
	genValidators1[0] = types.NewValidator(sdk.ValAddress(pk.Address()), pk, types.NewDescription("", "", "", "", ""))
	genValidators1[0].Tokens = sdk.OneInt()
	genValidators1[0].DelegatorShares = sdk.OneDec()

	tests := []struct {
		name    string
		mutate  func(*types.GenesisState)
		wantErr bool
	}{
		{"default", func(*types.GenesisState) {}, false},
		// validate genesis validators
		{"duplicate validator", func(data *types.GenesisState) {
			data.Validators = genValidators1
			data.Validators = append(data.Validators, genValidators1[0])
		}, true},
		{"no delegator shares", func(data *types.GenesisState) {
			data.Validators = genValidators1
			data.Validators[0].DelegatorShares = sdk.ZeroDec()
		}, true},
		{"jailed and bonded validator", func(data *types.GenesisState) {
			data.Validators = genValidators1
			data.Validators[0].Jailed = true
			data.Validators[0].Status = sdk.Bonded
		}, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			genesisState := types.DefaultGenesisState()
			tt.mutate(genesisState)
			if tt.wantErr {
				assert.Error(t, staking.ValidateGenesis(genesisState))
			} else {
				assert.NoError(t, staking.ValidateGenesis(genesisState))
			}
		})
	}
}

// ValidateGeneisTestSuite is a test suite to be used to test the ValidateAccountParamsInGenesis function
type ValidateGenesisTestSuite struct {
	suite.Suite

	ctx            sdk.Context
	app            *simapp.SimApp
	encodingConfig simappparams.EncodingConfig
}

func (suite *ValidateGenesisTestSuite) SetupTest() {
	checkTx := false
	app := simapp.Setup(checkTx)
	suite.ctx = app.BaseApp.NewContext(checkTx, tmproto.Header{})
	suite.app = app

	suite.encodingConfig = simapp.MakeEncodingConfig()
}

func (suite *ValidateGenesisTestSuite) setAccountBalance(addr sdk.AccAddress, amount int64) json.RawMessage {
	acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr)
	suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

	err := suite.app.BankKeeper.SetBalances(
		suite.ctx, addr, sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 25)},
	)
	suite.Require().NoError(err)

	bankGenesisState := suite.app.BankKeeper.ExportGenesis(suite.ctx)
	bankGenesis, err := suite.encodingConfig.Amino.MarshalJSON(bankGenesisState) // TODO switch this to use Marshaler
	suite.Require().NoError(err)

	return bankGenesis
}

func (suite *ValidateGenesisTestSuite) TestValidateAccountParamsInGenesis() {
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
			cdc := suite.encodingConfig.Marshaler

			suite.app.StakingKeeper.SetParams(suite.ctx, types.DefaultParams())
			stakingGenesisState := staking.ExportGenesis(suite.ctx, suite.app.StakingKeeper)
			suite.Require().Equal(stakingGenesisState.Params, types.DefaultParams())
			stakingGenesis, err := cdc.MarshalJSON(stakingGenesisState) // TODO switch this to use Marshaler
			suite.Require().NoError(err)
			appGenesisState[types.ModuleName] = stakingGenesis

			tc.malleate()
			err = staking.ValidateAccountParamsInGenesis(
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

func TestValidateGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(ValidateGenesisTestSuite))
}
