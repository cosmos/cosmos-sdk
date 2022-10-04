package keeper_test

import (
	"sort"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"
	"pgregory.net/rapid"

	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtestutil "github.com/cosmos/cosmos-sdk/x/staking/testutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type DeterministicTestSuite struct {
	suite.Suite

	ctx           sdk.Context
	stakingKeeper *stakingkeeper.Keeper
	bankKeeper    *stakingtestutil.MockBankKeeper
	accountKeeper *stakingtestutil.MockAccountKeeper
	queryClient   stakingtypes.QueryClient
}

func (s *DeterministicTestSuite) SetupTest() {
	key := sdk.NewKVStoreKey(stakingtypes.StoreKey)
	testCtx := testutil.DefaultContextWithDB(s.T(), key, sdk.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx.WithBlockHeader(tmproto.Header{Time: tmtime.Now()})
	encCfg := moduletestutil.MakeTestEncodingConfig()

	ctrl := gomock.NewController(s.T())
	accountKeeper := stakingtestutil.NewMockAccountKeeper(ctrl)
	accountKeeper.EXPECT().GetModuleAddress(stakingtypes.BondedPoolName).Return(bondedAcc.GetAddress())
	accountKeeper.EXPECT().GetModuleAddress(stakingtypes.NotBondedPoolName).Return(notBondedAcc.GetAddress())
	bankKeeper := stakingtestutil.NewMockBankKeeper(ctrl)

	keeper := stakingkeeper.NewKeeper(
		encCfg.Codec,
		key,
		accountKeeper,
		bankKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	keeper.SetParams(ctx, stakingtypes.DefaultParams())

	s.ctx = ctx
	s.stakingKeeper = keeper
	s.bankKeeper = bankKeeper
	s.accountKeeper = accountKeeper

	stakingtypes.RegisterInterfaces(encCfg.InterfaceRegistry)
	queryHelper := baseapp.NewQueryServerTestHelper(ctx, encCfg.InterfaceRegistry)
	stakingtypes.RegisterQueryServer(queryHelper, stakingkeeper.Querier{Keeper: keeper})
	s.queryClient = stakingtypes.NewQueryClient(queryHelper)
}

func drawDuration() *rapid.Generator[time.Duration] {
	return rapid.Custom(func(t *rapid.T) time.Duration {
		now := time.Now()
		// range from current time to 365days.
		duration := rapid.Int64Range(now.Unix(), 365*24*60*60*now.Unix()).Draw(t, "time")
		return time.Duration(duration)
	})
}

func (suite *DeterministicTestSuite) runParamsIterations(prevParams stakingtypes.Params) {
	for i := 0; i < 1000; i++ {
		res, err := suite.queryClient.Params(suite.ctx, &stakingtypes.QueryParamsRequest{})
		suite.Require().NoError(err)
		suite.Require().NotNil(res)

		suite.Require().Equal(res.Params, prevParams)
		prevParams = res.Params
	}
}

func TestDeterministicTestSuite(t *testing.T) {
	suite.Run(t, new(DeterministicTestSuite))
}

func (suite *DeterministicTestSuite) TestGRPCParams() {
	rapid.Check(suite.T(), func(t *rapid.T) {
		params := stakingtypes.Params{
			BondDenom:         rapid.StringMatching(sdk.DefaultCoinDenomRegex()).Draw(t, "bond-denom"),
			UnbondingTime:     drawDuration().Draw(t, "duration"),
			MaxValidators:     rapid.Uint32Min(1).Draw(t, "max-validators"),
			MaxEntries:        rapid.Uint32Min(1).Draw(t, "max-entries"),
			HistoricalEntries: rapid.Uint32Min(1).Draw(t, "historical-entries"),
			MinCommissionRate: sdk.NewDecWithPrec(rapid.Int64Range(0, 100).Draw(t, "commission"), 2),
		}

		err := suite.stakingKeeper.SetParams(suite.ctx, params)
		suite.Require().NoError(err)

		suite.runParamsIterations(params)
	})

	params := stakingtypes.Params{
		BondDenom:         "denom",
		UnbondingTime:     time.Hour,
		MaxValidators:     85,
		MaxEntries:        5,
		HistoricalEntries: 5,
		MinCommissionRate: sdk.NewDecWithPrec(5, 2),
	}

	err := suite.stakingKeeper.SetParams(suite.ctx, params)
	suite.Require().NoError(err)

	suite.runParamsIterations(params)
}

func drawPubKey() *rapid.Generator[ed25519.PubKey] {
	return rapid.Custom(func(t *rapid.T) ed25519.PubKey {
		pkBz := rapid.SliceOfN(rapid.Byte(), 32, 32).Draw(t, "hex")
		return ed25519.PubKey{Key: pkBz}
	})
}

func (suite *DeterministicTestSuite) matchValidators(val1, val2 stakingtypes.Validator) {
	suite.Require().Equal(val1.GetOperator(), val2.GetOperator())
	suite.Require().Equal(val1.ConsensusPubkey.Value, val2.ConsensusPubkey.Value)
	suite.Require().Equal(val1.Jailed, val2.Jailed)
	suite.Require().Equal(val1.GetStatus(), val2.GetStatus())
	suite.Require().Equal(val1.GetTokens(), val2.GetTokens())
	suite.Require().Equal(val1.GetDelegatorShares(), val2.GetDelegatorShares())
	suite.Require().Equal(val1.Description, val2.Description)
	suite.Require().Equal(val1.UnbondingHeight, val2.UnbondingHeight)
	suite.Require().True(val1.UnbondingTime.Equal(val2.UnbondingTime))
	suite.Require().Equal(val1.Commission, val2.Commission)
	suite.Require().Equal(val1.MinSelfDelegation, val2.MinSelfDelegation)
}

func (suite *DeterministicTestSuite) getValidator(t *rapid.T) stakingtypes.Validator {
	bond_types := []stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Unbonded, stakingtypes.Unbonding}

	pubkey := drawPubKey().Draw(t, "pubkey")
	pubkeyAny, err := codectypes.NewAnyWithValue(&pubkey)
	suite.Require().NoError(err)
	return stakingtypes.Validator{
		OperatorAddress: sdk.ValAddress(testdata.AddressGenerator(t).Draw(t, "address")).String(),
		ConsensusPubkey: pubkeyAny,
		Jailed:          rapid.Bool().Draw(t, "jailed"),
		Status:          bond_types[rapid.IntRange(0, 2).Draw(t, "bond-status")],
		Tokens:          sdk.NewInt(rapid.Int64Min(1).Draw(t, "tokens")),
		DelegatorShares: sdk.NewDecWithPrec(rapid.Int64Range(0, 100).Draw(t, "commission"), 2),
		Description: stakingtypes.NewDescription(
			rapid.StringN(5, 250, 255).Draw(t, "moniker"),
			rapid.StringN(5, 250, 255).Draw(t, "identity"),
			rapid.StringN(5, 250, 255).Draw(t, "website"),
			rapid.StringN(5, 250, 255).Draw(t, "securityContact"),
			rapid.StringN(5, 250, 255).Draw(t, "details"),
		),
		UnbondingHeight: rapid.Int64Min(1).Draw(t, "unbonding-height"),
		UnbondingTime:   time.Now().Add(drawDuration().Draw(t, "duration")),
		Commission: stakingtypes.NewCommission(
			sdk.NewDecWithPrec(rapid.Int64Range(0, 100).Draw(t, "rate"), 2),
			sdk.NewDecWithPrec(rapid.Int64Range(0, 100).Draw(t, "max-rate"), 2),
			sdk.NewDecWithPrec(rapid.Int64Range(0, 100).Draw(t, "max-change-rate"), 2),
		),
		MinSelfDelegation: sdk.NewInt(rapid.Int64Min(1).Draw(t, "tokens")),
	}
}

func (suite *DeterministicTestSuite) runValidatorIterations(valAddr string, prevValRes stakingtypes.Validator) {
	for i := 0; i < 1000; i++ {
		res, err := suite.queryClient.Validator(suite.ctx, &stakingtypes.QueryValidatorRequest{
			ValidatorAddr: valAddr,
		})

		suite.Require().NoError(err)
		suite.Require().NotNil(res)

		suite.matchValidators(res.GetValidator(), prevValRes)
		prevValRes = res.GetValidator()
	}
}

func (suite *DeterministicTestSuite) TestGRPCValidator() {
	rapid.Check(suite.T(), func(t *rapid.T) {
		val := suite.getValidator(t)
		suite.stakingKeeper.SetValidator(suite.ctx, val)

		suite.runValidatorIterations(val.OperatorAddress, val)
	})

	pubkey := ed25519.PubKey{Key: []byte{24, 179, 242, 2, 151, 3, 34, 6, 1, 11, 0, 194, 202, 201, 77, 1, 167, 40, 249, 115, 32, 97, 18, 1, 1, 127, 255, 103, 13, 1, 34, 1}}
	pubkeyAny, err := codectypes.NewAnyWithValue(&pubkey)
	suite.Require().NoError(err)

	val := stakingtypes.Validator{
		OperatorAddress: "cosmosvaloper1qqqryrs09ggeuqszqygqyqd2tgqmsqzewacjj7",
		ConsensusPubkey: pubkeyAny,
		Jailed:          false,
		Status:          stakingtypes.Bonded,
		Tokens:          sdk.NewInt(100),
		DelegatorShares: sdk.NewDecWithPrec(5, 2),
		Description: stakingtypes.NewDescription(
			"moniker",
			"identity",
			"website",
			"securityContact",
			"details",
		),
		UnbondingHeight: 10,
		UnbondingTime:   time.Date(2022, 10, 1, 0, 0, 0, 0, time.UTC),
		Commission: stakingtypes.NewCommission(
			sdk.NewDecWithPrec(5, 2),
			sdk.NewDecWithPrec(5, 2),
			sdk.NewDecWithPrec(5, 2),
		),
		MinSelfDelegation: sdk.NewInt(10),
	}

	suite.stakingKeeper.SetValidator(suite.ctx, val)
	suite.runValidatorIterations(val.OperatorAddress, val)
}

func (suite *DeterministicTestSuite) runValidatorsIterations(prevValsRes []stakingtypes.Validator) {
	for i := 0; i < 1000; i++ {
		res, err := suite.queryClient.Validators(suite.ctx, &stakingtypes.QueryValidatorsRequest{})

		suite.Require().NoError(err)
		suite.Require().NotNil(res)
		if i == 0 {
			temp := make([]stakingtypes.Validator, len(res.GetValidators()))
			copy(temp, res.GetValidators())

			sort.Slice(prevValsRes, func(i2, j int) bool {
				return prevValsRes[i2].OperatorAddress > prevValsRes[j].OperatorAddress
			})

			sort.Slice(temp, func(i2, j int) bool {
				return temp[i2].OperatorAddress > temp[j].OperatorAddress
			})

			for j := 0; j < len(temp); j++ {
				suite.matchValidators(temp[j], prevValsRes[j])
			}

			prevValsRes = res.GetValidators()
			continue
		}

		for j := 0; j < len(prevValsRes); j++ {
			suite.matchValidators(res.GetValidators()[j], prevValsRes[j])
		}

		prevValsRes = res.GetValidators()
	}
}

func (suite *DeterministicTestSuite) getStaticValidator() stakingtypes.Validator {
	pubkey := ed25519.PubKey{Key: []byte{24, 179, 242, 2, 151, 3, 34, 6, 1, 11, 0, 194, 202, 201, 77, 1, 167, 40, 249, 115, 32, 97, 18, 1, 1, 127, 255, 103, 13, 1, 34, 1}}
	pubkeyAny, err := codectypes.NewAnyWithValue(&pubkey)
	suite.Require().NoError(err)

	val := stakingtypes.Validator{
		OperatorAddress: "cosmosvaloper1qqqryrs09ggeuqszqygqyqd2tgqmsqzewacjj7",
		ConsensusPubkey: pubkeyAny,
		Jailed:          false,
		Status:          stakingtypes.Bonded,
		Tokens:          sdk.NewInt(100),
		DelegatorShares: sdk.NewDecWithPrec(5, 2),
		Description: stakingtypes.NewDescription(
			"moniker",
			"identity",
			"website",
			"securityContact",
			"details",
		),
		UnbondingHeight: 10,
		UnbondingTime:   time.Date(2022, 10, 1, 0, 0, 0, 0, time.UTC),
		Commission: stakingtypes.NewCommission(
			sdk.NewDecWithPrec(5, 2),
			sdk.NewDecWithPrec(5, 2),
			sdk.NewDecWithPrec(5, 2),
		),
		MinSelfDelegation: sdk.NewInt(10),
	}

	return val
}

func (suite *DeterministicTestSuite) TestGRPCValidators() {

	rapid.Check(suite.T(), func(t *rapid.T) {
		suite.SetupTest() // reset

		valsCount := rapid.IntRange(1, 5).Draw(t, "num-validators")
		vals := make([]stakingtypes.Validator, 0, valsCount)

		for i := 0; i < valsCount; i++ {
			val := suite.getValidator(t)
			suite.stakingKeeper.SetValidator(suite.ctx, val)
			vals = append(vals, val)
		}

		suite.runValidatorsIterations(vals)
	})

	suite.SetupTest() // reset
	val1 := suite.getStaticValidator()

	pubkey2 := ed25519.PubKey{Key: []byte{40, 249, 115, 32, 97, 18, 1, 1, 127, 255, 103, 13, 1, 34, 1, 24, 179, 242, 2, 151, 3, 34, 6, 1, 11, 0, 194, 202, 201, 77, 1, 167}}
	pubkeyAny2, err := codectypes.NewAnyWithValue(&pubkey2)
	suite.Require().NoError(err)

	val2 := stakingtypes.Validator{
		OperatorAddress: "cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj",
		ConsensusPubkey: pubkeyAny2,
		Jailed:          true,
		Status:          stakingtypes.Bonded,
		Tokens:          sdk.NewInt(10012),
		DelegatorShares: sdk.NewDecWithPrec(96, 2),
		Description: stakingtypes.NewDescription(
			"moniker",
			"identity",
			"website",
			"securityContact",
			"details",
		),
		UnbondingHeight: 100132,
		UnbondingTime:   time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC),
		Commission: stakingtypes.NewCommission(
			sdk.NewDecWithPrec(15, 2),
			sdk.NewDecWithPrec(59, 2),
			sdk.NewDecWithPrec(51, 2),
		),
		MinSelfDelegation: sdk.NewInt(1),
	}

	suite.stakingKeeper.SetValidator(suite.ctx, val1)
	suite.stakingKeeper.SetValidator(suite.ctx, val2)

	suite.runValidatorsIterations([]stakingtypes.Validator{val1, val2})
}

func (suite *DeterministicTestSuite) runValidatorDelegationsIterations(validator sdk.ValAddress, prevDels stakingtypes.DelegationResponses) {
	for i := 0; i < 1000; i++ {
		res, err := suite.queryClient.ValidatorDelegations(suite.ctx, &stakingtypes.QueryValidatorDelegationsRequest{
			ValidatorAddr: validator.String(),
		})
		suite.Require().NoError(err)

		if i == 0 {
			temp := make(stakingtypes.DelegationResponses, len(res.DelegationResponses))
			copy(temp, res.DelegationResponses)

			sort.Slice(temp, func(i2, j int) bool {
				return temp[i2].Delegation.DelegatorAddress > temp[j].Delegation.DelegatorAddress
			})

			sort.Slice(prevDels, func(i2, j int) bool {
				return prevDels[i2].Delegation.DelegatorAddress > prevDels[j].Delegation.DelegatorAddress
			})

			suite.Require().Equal(temp, prevDels)
			prevDels = res.DelegationResponses
			continue
		}

		suite.Require().Equal(res.DelegationResponses, prevDels)
		prevDels = res.DelegationResponses
	}
}

func (suite *DeterministicTestSuite) TestGRPCValidatorDelegations() {

	rapid.Check(suite.T(), func(t *rapid.T) {
		suite.SetupTest() // reset
		validator := suite.getValidator(t)

		numDels := rapid.IntRange(1, 5).Draw(t, "num-dels")
		delegations := make([]stakingtypes.Delegation, 0, numDels)
		delegationResponses := make(stakingtypes.DelegationResponses, 0, numDels)

		moduleName := stakingtypes.NotBondedPoolName
		if validator.GetStatus() == stakingtypes.Bonded {
			moduleName = stakingtypes.BondedPoolName
		}

		suite.stakingKeeper.SetValidator(suite.ctx, validator)
		suite.stakingKeeper.SetValidatorByPowerIndex(suite.ctx, validator)

		for i := 0; i < numDels; i++ {
			delegator := testdata.AddressGenerator(t).Draw(t, "delegator")
			amt := suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, rapid.Int64Range(100, 1000).Draw(t, "amount"))

			suite.bankKeeper.EXPECT().DelegateCoinsFromAccountToModule(
				suite.ctx, delegator, moduleName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, amt))).Return(nil)

			newShares, err := suite.stakingKeeper.Delegate(suite.ctx, delegator, amt, stakingtypes.Unbonded, validator, true)
			suite.Require().NoError(err)

			delegation := stakingtypes.NewDelegation(delegator, validator.GetOperator(), newShares)
			suite.stakingKeeper.SetDelegation(suite.ctx, delegation)

			delegations = append(delegations, delegation)
			delegationResponses, err = stakingkeeper.DelegationsToDelegationResponses(suite.ctx, suite.stakingKeeper, delegations)
			suite.Require().NoError(err)
		}

		suite.runValidatorDelegationsIterations(validator.GetOperator(), delegationResponses)
	})

	suite.SetupTest() // reset

	validator := suite.getStaticValidator()
	valAddr, err := sdk.ValAddressFromBech32(validator.OperatorAddress)
	suite.Require().NoError(err)

	moduleName := stakingtypes.NotBondedPoolName

	delegations := make([]stakingtypes.Delegation, 0, 2)

	if validator.GetStatus() == stakingtypes.Bonded {
		moduleName = stakingtypes.BondedPoolName
	}

	suite.stakingKeeper.SetValidator(suite.ctx, validator)
	suite.stakingKeeper.SetValidatorByPowerIndex(suite.ctx, validator)

	delegator1 := sdk.MustAccAddressFromBech32("cosmos1nph3cfzk6trsmfxkeu943nvach5qw4vwstnvkl")
	amt1 := suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, 101)

	suite.bankKeeper.EXPECT().DelegateCoinsFromAccountToModule(
		suite.ctx, delegator1, moduleName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, amt1))).Return(nil)

	newShares1, err := suite.stakingKeeper.Delegate(suite.ctx, delegator1, amt1, stakingtypes.Unbonded, validator, true)
	suite.Require().NoError(err)

	delegator2 := sdk.MustAccAddressFromBech32("cosmos139f7kncmglres2nf3h4hc4tade85ekfr8sulz5")
	amt2 := suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, 102)
	suite.bankKeeper.EXPECT().DelegateCoinsFromAccountToModule(
		suite.ctx, delegator2, moduleName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, amt2))).Return(nil)
	newShares2, err := suite.stakingKeeper.Delegate(suite.ctx, delegator2, amt2, stakingtypes.Unbonded, validator, true)
	suite.Require().NoError(err)

	delegations = append(delegations, stakingtypes.NewDelegation(delegator1, valAddr, newShares1),
		stakingtypes.NewDelegation(delegator2, valAddr, newShares2))

	delegationResps, err := stakingkeeper.DelegationsToDelegationResponses(suite.ctx, suite.stakingKeeper, delegations)
	suite.Require().NoError(err)

	suite.runValidatorDelegationsIterations(validator.GetOperator(), delegationResps)
}
