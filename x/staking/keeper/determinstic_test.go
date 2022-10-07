package keeper_test

import (
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

func (suite *DeterministicTestSuite) getValidatorWithStatus(t *rapid.T, status stakingtypes.BondStatus) stakingtypes.Validator {
	val := suite.getValidator(t)
	val.Status = status
	return val
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

func (suite *DeterministicTestSuite) runValidatorsIterations(req *stakingtypes.QueryValidatorsRequest, prevRes *stakingtypes.QueryValidatorsResponse) {
	for i := 0; i < 1000; i++ {
		res, err := suite.queryClient.Validators(suite.ctx, req)

		suite.Require().NoError(err)
		suite.Require().NotNil(res)

		suite.Require().Equal(res, prevRes)
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
	validatorStatus := []string{stakingtypes.BondStatusBonded, stakingtypes.BondStatusUnbonded, stakingtypes.BondStatusUnbonding, ""}
	rapid.Check(suite.T(), func(t *rapid.T) {
		suite.SetupTest() // reset

		valsCount := rapid.IntRange(1, 20).Draw(t, "num-validators")

		for i := 0; i < valsCount; i++ {
			val := suite.getValidator(t)
			suite.stakingKeeper.SetValidator(suite.ctx, val)
		}

		req := &stakingtypes.QueryValidatorsRequest{
			Status:     validatorStatus[rapid.IntRange(0, 3).Draw(t, "status")],
			Pagination: testdata.PaginationGenerator(t, uint64(valsCount)).Draw(t, "pagination"),
		}
		res, err := suite.queryClient.Validators(suite.ctx, req)
		suite.Require().NoError(err)

		suite.runValidatorsIterations(req, res)
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

	req := &stakingtypes.QueryValidatorsRequest{}
	res, err := suite.queryClient.Validators(suite.ctx, req)
	suite.Require().NoError(err)

	suite.runValidatorsIterations(req, res)
}

func (suite *DeterministicTestSuite) runValidatorDelegationsIterations(req *stakingtypes.QueryValidatorDelegationsRequest, prevDels *stakingtypes.QueryValidatorDelegationsResponse) {
	for i := 0; i < 1000; i++ {
		res, err := suite.queryClient.ValidatorDelegations(suite.ctx, req)
		suite.Require().NoError(err)

		suite.Require().Equal(res, prevDels)
	}
}

func (suite *DeterministicTestSuite) TestGRPCValidatorDelegations() {

	rapid.Check(suite.T(), func(t *rapid.T) {
		suite.SetupTest() // reset
		validator := suite.getValidator(t)

		numDels := rapid.IntRange(1, 5).Draw(t, "num-dels")
		moduleName := stakingtypes.NotBondedPoolName
		if validator.GetStatus() == stakingtypes.Bonded {
			moduleName = stakingtypes.BondedPoolName
		}

		suite.stakingKeeper.SetValidator(suite.ctx, validator)
		suite.stakingKeeper.SetValidatorByPowerIndex(suite.ctx, validator)

		for i := 0; i < numDels; i++ {
			delegator := testdata.AddressGenerator(t).Draw(t, "delegator")
			amt := suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, rapid.Int64Range(100, 1000).Draw(t, "amount"))

			// TODO remove mocks
			suite.bankKeeper.EXPECT().DelegateCoinsFromAccountToModule(
				suite.ctx, delegator, moduleName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, amt))).Return(nil)

			_, err := suite.stakingKeeper.Delegate(suite.ctx, delegator, amt, stakingtypes.Unbonded, validator, true)
			suite.Require().NoError(err)
		}

		req := &stakingtypes.QueryValidatorDelegationsRequest{
			ValidatorAddr: validator.OperatorAddress,
			Pagination:    testdata.PaginationGenerator(t, uint64(numDels)).Draw(t, "pagination"),
		}

		res, err := suite.queryClient.ValidatorDelegations(suite.ctx, req)
		suite.Require().NoError(err)
		suite.runValidatorDelegationsIterations(req, res)
	})

	suite.SetupTest() // reset

	validator := suite.getStaticValidator()
	moduleName := stakingtypes.NotBondedPoolName

	if validator.GetStatus() == stakingtypes.Bonded {
		moduleName = stakingtypes.BondedPoolName
	}

	suite.stakingKeeper.SetValidator(suite.ctx, validator)
	suite.stakingKeeper.SetValidatorByPowerIndex(suite.ctx, validator)

	delegator1 := sdk.MustAccAddressFromBech32("cosmos1nph3cfzk6trsmfxkeu943nvach5qw4vwstnvkl")
	amt1 := suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, 101)

	suite.bankKeeper.EXPECT().DelegateCoinsFromAccountToModule(
		suite.ctx, delegator1, moduleName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, amt1))).Return(nil)

	_, err := suite.stakingKeeper.Delegate(suite.ctx, delegator1, amt1, stakingtypes.Unbonded, validator, true)
	suite.Require().NoError(err)

	delegator2 := sdk.MustAccAddressFromBech32("cosmos139f7kncmglres2nf3h4hc4tade85ekfr8sulz5")
	amt2 := suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, 102)
	suite.bankKeeper.EXPECT().DelegateCoinsFromAccountToModule(
		suite.ctx, delegator2, moduleName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, amt2))).Return(nil)
	_, err = suite.stakingKeeper.Delegate(suite.ctx, delegator2, amt2, stakingtypes.Unbonded, validator, true)
	suite.Require().NoError(err)

	req := &stakingtypes.QueryValidatorDelegationsRequest{
		ValidatorAddr: validator.OperatorAddress,
	}

	res, err := suite.queryClient.ValidatorDelegations(suite.ctx, req)
	suite.Require().NoError(err)
	suite.runValidatorDelegationsIterations(req, res)
}

func (suite *DeterministicTestSuite) runValidatorUnbondingDelegationsIterations(req *stakingtypes.QueryValidatorUnbondingDelegationsRequest, prevRes *stakingtypes.QueryValidatorUnbondingDelegationsResponse) {
	for i := 0; i < 1000; i++ {
		res, err := suite.queryClient.ValidatorUnbondingDelegations(suite.ctx, req)
		suite.Require().NoError(err)
		suite.Require().Equal(res, prevRes)
	}
}

func (suite *DeterministicTestSuite) TestGRPCValidatorUnbondingDelegations() {

	rapid.Check(suite.T(), func(t *rapid.T) {
		suite.SetupTest() // reset
		validator := suite.getValidatorWithStatus(t, stakingtypes.Bonded)
		numDels := rapid.IntRange(1, 5).Draw(t, "num-dels")

		suite.stakingKeeper.SetValidator(suite.ctx, validator)
		suite.stakingKeeper.SetValidatorByPowerIndex(suite.ctx, validator)

		for i := 0; i < numDels; i++ {
			delegator := testdata.AddressGenerator(t).Draw(t, "delegator")
			amt := suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, rapid.Int64Range(100, 1000).Draw(t, "amount"))

			suite.bankKeeper.EXPECT().DelegateCoinsFromAccountToModule(
				suite.ctx, delegator, stakingtypes.BondedPoolName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, amt))).Return(nil)
			newShares, err := suite.stakingKeeper.Delegate(suite.ctx, delegator, amt, stakingtypes.Unbonded, validator, true)
			suite.Require().NoError(err)

			suite.bankKeeper.EXPECT().SendCoinsFromModuleToModule(
				gomock.Any(), stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName, gomock.Any()).Return(nil)

			_, err = suite.stakingKeeper.Undelegate(suite.ctx, delegator, validator.GetOperator(), newShares)
			suite.Require().NoError(err)
		}

		req := &stakingtypes.QueryValidatorUnbondingDelegationsRequest{
			ValidatorAddr: validator.OperatorAddress,
			Pagination:    testdata.PaginationGenerator(t, uint64(numDels)).Draw(t, "pagination"),
		}
		res, err := suite.queryClient.ValidatorUnbondingDelegations(suite.ctx, req)
		suite.Require().NoError(err)

		suite.runValidatorUnbondingDelegationsIterations(req, res)
	})

	suite.SetupTest() // reset

	validator := suite.getStaticValidator()
	valAddr, err := sdk.ValAddressFromBech32(validator.OperatorAddress)
	suite.Require().NoError(err)

	suite.stakingKeeper.SetValidator(suite.ctx, validator)
	suite.stakingKeeper.SetValidatorByPowerIndex(suite.ctx, validator)

	delegator1 := sdk.MustAccAddressFromBech32("cosmos1nph3cfzk6trsmfxkeu943nvach5qw4vwstnvkl")
	amt1 := suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, 101)

	suite.bankKeeper.EXPECT().DelegateCoinsFromAccountToModule(
		suite.ctx, delegator1, stakingtypes.BondedPoolName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, amt1))).Return(nil)

	newShares1, err := suite.stakingKeeper.Delegate(suite.ctx, delegator1, amt1, stakingtypes.Unbonded, validator, true)
	suite.Require().NoError(err)

	suite.bankKeeper.EXPECT().SendCoinsFromModuleToModule(
		gomock.Any(), stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName, gomock.Any()).Return(nil)
	_, err = suite.stakingKeeper.Undelegate(suite.ctx, delegator1, valAddr, newShares1)
	suite.Require().NoError(err)

	delegator2 := sdk.MustAccAddressFromBech32("cosmos139f7kncmglres2nf3h4hc4tade85ekfr8sulz5")
	amt2 := suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, 102)

	suite.bankKeeper.EXPECT().DelegateCoinsFromAccountToModule(
		suite.ctx, delegator2, stakingtypes.BondedPoolName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, amt2))).Return(nil)

	newShares2, err := suite.stakingKeeper.Delegate(suite.ctx, delegator2, amt2, stakingtypes.Unbonded, validator, true)
	suite.Require().NoError(err)

	suite.bankKeeper.EXPECT().SendCoinsFromModuleToModule(
		gomock.Any(), stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName, gomock.Any()).Return(nil)
	_, err = suite.stakingKeeper.Undelegate(suite.ctx, delegator2, valAddr, newShares2)
	suite.Require().NoError(err)

	req := &stakingtypes.QueryValidatorUnbondingDelegationsRequest{
		ValidatorAddr: validator.OperatorAddress,
	}
	res, err := suite.queryClient.ValidatorUnbondingDelegations(suite.ctx, req)
	suite.Require().NoError(err)

	suite.runValidatorUnbondingDelegationsIterations(req, res)
}

func (suite *DeterministicTestSuite) runDelegationIteratons(req *stakingtypes.QueryDelegationRequest, prevRes *stakingtypes.QueryDelegationResponse) {
	for i := 0; i < 1000; i++ {
		res, err := suite.queryClient.Delegation(suite.ctx, req)
		suite.Require().NoError(err)
		suite.Require().Equal(res, prevRes)
	}
}

func (suite *DeterministicTestSuite) TestGRPCDelegation() {
	rapid.Check(suite.T(), func(t *rapid.T) {
		validator := suite.getValidatorWithStatus(t, stakingtypes.Bonded)

		suite.stakingKeeper.SetValidator(suite.ctx, validator)
		suite.stakingKeeper.SetValidatorByPowerIndex(suite.ctx, validator)

		delegator := testdata.AddressGenerator(t).Draw(t, "delegator")
		amt := suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, rapid.Int64Range(100, 1000).Draw(t, "amount"))

		// TODO remove mocks
		suite.bankKeeper.EXPECT().DelegateCoinsFromAccountToModule(
			suite.ctx, delegator, stakingtypes.BondedPoolName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, amt))).Return(nil)

		_, err := suite.stakingKeeper.Delegate(suite.ctx, delegator, amt, stakingtypes.Unbonded, validator, true)
		suite.Require().NoError(err)

		req := &stakingtypes.QueryDelegationRequest{
			ValidatorAddr: validator.OperatorAddress,
			DelegatorAddr: delegator.String(),
		}

		res, err := suite.queryClient.Delegation(suite.ctx, req)
		suite.Require().NoError(err)
		suite.runDelegationIteratons(req, res)
	})

	suite.SetupTest() // reset

	validator := suite.getStaticValidator()

	suite.stakingKeeper.SetValidator(suite.ctx, validator)
	suite.stakingKeeper.SetValidatorByPowerIndex(suite.ctx, validator)

	delegator1 := sdk.MustAccAddressFromBech32("cosmos1nph3cfzk6trsmfxkeu943nvach5qw4vwstnvkl")

	amt1 := suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, 101)

	suite.bankKeeper.EXPECT().DelegateCoinsFromAccountToModule(
		suite.ctx, delegator1, stakingtypes.BondedPoolName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, amt1))).Return(nil)

	_, err := suite.stakingKeeper.Delegate(suite.ctx, delegator1, amt1, stakingtypes.Unbonded, validator, true)
	suite.Require().NoError(err)

	req := &stakingtypes.QueryDelegationRequest{
		ValidatorAddr: validator.OperatorAddress,
		DelegatorAddr: delegator1.String(),
	}

	res, err := suite.queryClient.Delegation(suite.ctx, req)
	suite.Require().NoError(err)
	suite.runDelegationIteratons(req, res)
}

func (suite *DeterministicTestSuite) runUnbondingDelegationIterations(req *stakingtypes.QueryUnbondingDelegationRequest, prevRes *stakingtypes.QueryUnbondingDelegationResponse) {
	for i := 0; i < 1000; i++ {
		res, err := suite.queryClient.UnbondingDelegation(suite.ctx, req)
		suite.Require().NoError(err)
		suite.Require().Equal(res, prevRes)
	}
}

func (suite *DeterministicTestSuite) TestGRPCUnbondingDelegation() {

	rapid.Check(suite.T(), func(t *rapid.T) {
		suite.SetupTest() // reset
		validator := suite.getValidatorWithStatus(t, stakingtypes.Bonded)

		suite.stakingKeeper.SetValidator(suite.ctx, validator)
		suite.stakingKeeper.SetValidatorByPowerIndex(suite.ctx, validator)

		delegator := testdata.AddressGenerator(t).Draw(t, "delegator")
		amt := suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, rapid.Int64Range(100, 1000).Draw(t, "amount"))

		suite.bankKeeper.EXPECT().DelegateCoinsFromAccountToModule(
			suite.ctx, delegator, stakingtypes.BondedPoolName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, amt))).Return(nil)
		newShares, err := suite.stakingKeeper.Delegate(suite.ctx, delegator, amt, stakingtypes.Unbonded, validator, true)
		suite.Require().NoError(err)

		suite.bankKeeper.EXPECT().SendCoinsFromModuleToModule(
			gomock.Any(), stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName, gomock.Any()).Return(nil)

		_, err = suite.stakingKeeper.Undelegate(suite.ctx, delegator, validator.GetOperator(), newShares)
		suite.Require().NoError(err)

		req := &stakingtypes.QueryUnbondingDelegationRequest{
			ValidatorAddr: validator.OperatorAddress,
			DelegatorAddr: delegator.String(),
		}
		res, err := suite.queryClient.UnbondingDelegation(suite.ctx, req)
		suite.Require().NoError(err)

		suite.runUnbondingDelegationIterations(req, res)
	})

	suite.SetupTest() // reset
	validator := suite.getStaticValidator()
	valAddr, err := sdk.ValAddressFromBech32(validator.OperatorAddress)
	suite.Require().NoError(err)

	suite.stakingKeeper.SetValidator(suite.ctx, validator)
	suite.stakingKeeper.SetValidatorByPowerIndex(suite.ctx, validator)

	delegator1 := sdk.MustAccAddressFromBech32("cosmos1nph3cfzk6trsmfxkeu943nvach5qw4vwstnvkl")
	amt1 := suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, 101)

	suite.bankKeeper.EXPECT().DelegateCoinsFromAccountToModule(
		suite.ctx, delegator1, stakingtypes.BondedPoolName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, amt1))).Return(nil)

	newShares1, err := suite.stakingKeeper.Delegate(suite.ctx, delegator1, amt1, stakingtypes.Unbonded, validator, true)
	suite.Require().NoError(err)

	suite.bankKeeper.EXPECT().SendCoinsFromModuleToModule(
		gomock.Any(), stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName, gomock.Any()).Return(nil)
	_, err = suite.stakingKeeper.Undelegate(suite.ctx, delegator1, valAddr, newShares1)
	suite.Require().NoError(err)

	req := &stakingtypes.QueryUnbondingDelegationRequest{
		ValidatorAddr: validator.OperatorAddress,
		DelegatorAddr: delegator1.String(),
	}
	res, err := suite.queryClient.UnbondingDelegation(suite.ctx, req)
	suite.Require().NoError(err)

	suite.runUnbondingDelegationIterations(req, res)
}

func (suite *DeterministicTestSuite) runDelegatorDelegationsIterations(req *stakingtypes.QueryDelegatorDelegationsRequest, prevDels *stakingtypes.QueryDelegatorDelegationsResponse) {
	for i := 0; i < 1000; i++ {
		res, err := suite.queryClient.DelegatorDelegations(suite.ctx, req)
		suite.Require().NoError(err)

		suite.Require().Equal(res, prevDels)
	}
}

func (suite *DeterministicTestSuite) TestGRPCDelegatorDelegations() {

	rapid.Check(suite.T(), func(t *rapid.T) {
		suite.SetupTest() // reset
		numVals := rapid.IntRange(1, 3).Draw(t, "num-dels")

		delegator := testdata.AddressGenerator(t).Draw(t, "delegator")

		for i := 0; i < numVals; i++ {
			amt := suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, rapid.Int64Range(100, 1000).Draw(t, "amount"))
			validator := suite.getValidatorWithStatus(t, stakingtypes.Bonded)
			suite.stakingKeeper.SetValidator(suite.ctx, validator)
			suite.stakingKeeper.SetValidatorByPowerIndex(suite.ctx, validator)

			// TODO remove mocks
			suite.bankKeeper.EXPECT().DelegateCoinsFromAccountToModule(
				suite.ctx, delegator, stakingtypes.BondedPoolName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, amt))).Return(nil)

			_, err := suite.stakingKeeper.Delegate(suite.ctx, delegator, amt, stakingtypes.Unbonded, validator, true)
			suite.Require().NoError(err)
		}

		req := &stakingtypes.QueryDelegatorDelegationsRequest{
			DelegatorAddr: delegator.String(),
			Pagination:    testdata.PaginationGenerator(t, uint64(numVals)).Draw(t, "pagination"),
		}

		res, err := suite.queryClient.DelegatorDelegations(suite.ctx, req)
		suite.Require().NoError(err)
		suite.runDelegatorDelegationsIterations(req, res)
	})

	suite.SetupTest() // reset

	validator := suite.getStaticValidator()
	moduleName := stakingtypes.NotBondedPoolName

	if validator.GetStatus() == stakingtypes.Bonded {
		moduleName = stakingtypes.BondedPoolName
	}

	suite.stakingKeeper.SetValidator(suite.ctx, validator)
	suite.stakingKeeper.SetValidatorByPowerIndex(suite.ctx, validator)

	delegator1 := sdk.MustAccAddressFromBech32("cosmos1nph3cfzk6trsmfxkeu943nvach5qw4vwstnvkl")
	amt1 := suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, 101)

	suite.bankKeeper.EXPECT().DelegateCoinsFromAccountToModule(
		suite.ctx, delegator1, moduleName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, amt1))).Return(nil)

	_, err := suite.stakingKeeper.Delegate(suite.ctx, delegator1, amt1, stakingtypes.Unbonded, validator, true)
	suite.Require().NoError(err)

	req := &stakingtypes.QueryDelegatorDelegationsRequest{
		DelegatorAddr: delegator1.String(),
	}

	res, err := suite.queryClient.DelegatorDelegations(suite.ctx, req)
	suite.Require().NoError(err)
	suite.runDelegatorDelegationsIterations(req, res)
}

func (suite *DeterministicTestSuite) runDelegatorValidatorIterations(req *stakingtypes.QueryDelegatorValidatorRequest, prevDels *stakingtypes.QueryDelegatorValidatorResponse) {
	for i := 0; i < 1000; i++ {
		res, err := suite.queryClient.DelegatorValidator(suite.ctx, req)
		suite.Require().NoError(err)

		suite.Require().Equal(res, prevDels)
	}
}

func (suite *DeterministicTestSuite) TestGRPCDelegatorValidator() {

	rapid.Check(suite.T(), func(t *rapid.T) {
		suite.SetupTest() // reset
		validator := suite.getValidatorWithStatus(t, stakingtypes.Bonded)

		suite.stakingKeeper.SetValidator(suite.ctx, validator)
		suite.stakingKeeper.SetValidatorByPowerIndex(suite.ctx, validator)

		delegator := testdata.AddressGenerator(t).Draw(t, "delegator")
		amt := suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, rapid.Int64Range(100, 1000).Draw(t, "amount"))

		// TODO remove mocks
		suite.bankKeeper.EXPECT().DelegateCoinsFromAccountToModule(
			suite.ctx, delegator, stakingtypes.BondedPoolName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, amt))).Return(nil)

		_, err := suite.stakingKeeper.Delegate(suite.ctx, delegator, amt, stakingtypes.Unbonded, validator, true)
		suite.Require().NoError(err)

		req := &stakingtypes.QueryDelegatorValidatorRequest{
			DelegatorAddr: delegator.String(),
			ValidatorAddr: validator.OperatorAddress,
		}

		res, err := suite.queryClient.DelegatorValidator(suite.ctx, req)
		suite.Require().NoError(err)
		suite.runDelegatorValidatorIterations(req, res)
	})

	suite.SetupTest() // reset

	validator := suite.getStaticValidator()
	moduleName := stakingtypes.NotBondedPoolName

	if validator.GetStatus() == stakingtypes.Bonded {
		moduleName = stakingtypes.BondedPoolName
	}

	suite.stakingKeeper.SetValidator(suite.ctx, validator)
	suite.stakingKeeper.SetValidatorByPowerIndex(suite.ctx, validator)

	delegator1 := sdk.MustAccAddressFromBech32("cosmos1nph3cfzk6trsmfxkeu943nvach5qw4vwstnvkl")
	amt1 := suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, 101)

	suite.bankKeeper.EXPECT().DelegateCoinsFromAccountToModule(
		suite.ctx, delegator1, moduleName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, amt1))).Return(nil)

	_, err := suite.stakingKeeper.Delegate(suite.ctx, delegator1, amt1, stakingtypes.Unbonded, validator, true)
	suite.Require().NoError(err)

	req := &stakingtypes.QueryDelegatorValidatorRequest{
		DelegatorAddr: delegator1.String(),
		ValidatorAddr: validator.OperatorAddress,
	}

	res, err := suite.queryClient.DelegatorValidator(suite.ctx, req)
	suite.Require().NoError(err)
	suite.runDelegatorValidatorIterations(req, res)
}

func (suite *DeterministicTestSuite) runDelegatorUnbondingDelegationsIterations(req *stakingtypes.QueryDelegatorUnbondingDelegationsRequest, prevRes *stakingtypes.QueryDelegatorUnbondingDelegationsResponse) {
	for i := 0; i < 1000; i++ {
		res, err := suite.queryClient.DelegatorUnbondingDelegations(suite.ctx, req)
		suite.Require().NoError(err)
		suite.Require().Equal(res, prevRes)
	}
}

func (suite *DeterministicTestSuite) TestGRPCDelegatorUnbondingDelegations() {

	rapid.Check(suite.T(), func(t *rapid.T) {
		suite.SetupTest() // reset
		numVals := rapid.IntRange(1, 5).Draw(t, "num-vals")
		delegator := testdata.AddressGenerator(t).Draw(t, "delegator")

		for i := 0; i < numVals; i++ {
			validator := suite.getValidatorWithStatus(t, stakingtypes.Bonded)

			suite.stakingKeeper.SetValidator(suite.ctx, validator)
			suite.stakingKeeper.SetValidatorByPowerIndex(suite.ctx, validator)
			amt := suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, rapid.Int64Range(100, 1000).Draw(t, "amount"))

			suite.bankKeeper.EXPECT().DelegateCoinsFromAccountToModule(
				suite.ctx, delegator, stakingtypes.BondedPoolName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, amt))).Return(nil)
			newShares, err := suite.stakingKeeper.Delegate(suite.ctx, delegator, amt, stakingtypes.Unbonded, validator, true)
			suite.Require().NoError(err)

			suite.bankKeeper.EXPECT().SendCoinsFromModuleToModule(
				gomock.Any(), stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName, gomock.Any()).Return(nil)

			_, err = suite.stakingKeeper.Undelegate(suite.ctx, delegator, validator.GetOperator(), newShares)
			suite.Require().NoError(err)
		}

		req := &stakingtypes.QueryDelegatorUnbondingDelegationsRequest{
			DelegatorAddr: delegator.String(),
			Pagination:    testdata.PaginationGenerator(t, uint64(numVals)).Draw(t, "pagination"),
		}
		res, err := suite.queryClient.DelegatorUnbondingDelegations(suite.ctx, req)
		suite.Require().NoError(err)

		suite.runDelegatorUnbondingDelegationsIterations(req, res)
	})

	suite.SetupTest() // reset

	validator := suite.getStaticValidator()
	valAddr, err := sdk.ValAddressFromBech32(validator.OperatorAddress)
	suite.Require().NoError(err)

	suite.stakingKeeper.SetValidator(suite.ctx, validator)
	suite.stakingKeeper.SetValidatorByPowerIndex(suite.ctx, validator)

	delegator1 := sdk.MustAccAddressFromBech32("cosmos1nph3cfzk6trsmfxkeu943nvach5qw4vwstnvkl")
	amt1 := suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, 101)

	suite.bankKeeper.EXPECT().DelegateCoinsFromAccountToModule(
		suite.ctx, delegator1, stakingtypes.BondedPoolName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, amt1))).Return(nil)

	newShares1, err := suite.stakingKeeper.Delegate(suite.ctx, delegator1, amt1, stakingtypes.Unbonded, validator, true)
	suite.Require().NoError(err)

	suite.bankKeeper.EXPECT().SendCoinsFromModuleToModule(
		gomock.Any(), stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName, gomock.Any()).Return(nil)
	_, err = suite.stakingKeeper.Undelegate(suite.ctx, delegator1, valAddr, newShares1)
	suite.Require().NoError(err)

	req := &stakingtypes.QueryDelegatorUnbondingDelegationsRequest{
		DelegatorAddr: delegator1.String(),
	}
	res, err := suite.queryClient.DelegatorUnbondingDelegations(suite.ctx, req)
	suite.Require().NoError(err)

	suite.runDelegatorUnbondingDelegationsIterations(req, res)
}

func (suite *DeterministicTestSuite) runHistoricalInfoIterations(req *stakingtypes.QueryHistoricalInfoRequest, prevRes *stakingtypes.QueryHistoricalInfoResponse) {
	for i := 0; i < 1000; i++ {
		res, err := suite.queryClient.HistoricalInfo(suite.ctx, req)
		suite.Require().NoError(err)
		suite.Require().Equal(res, prevRes)
	}
}

func (suite *DeterministicTestSuite) TestGRPCHistoricalInfo() {

	rapid.Check(suite.T(), func(t *rapid.T) {
		numVals := rapid.IntRange(1, 5).Draw(t, "num-vals")
		vals := make(stakingtypes.Validators, 0, numVals)
		for i := 0; i < numVals; i++ {
			validator := suite.getValidatorWithStatus(t, stakingtypes.Bonded)

			suite.stakingKeeper.SetValidator(suite.ctx, validator)
			suite.stakingKeeper.SetValidatorByPowerIndex(suite.ctx, validator)
			vals = append(vals, validator)
		}

		historicalInfo := stakingtypes.HistoricalInfo{
			Header: tmproto.Header{},
			Valset: vals,
		}

		height := rapid.Int64Min(0).Draw(t, "height")

		suite.stakingKeeper.SetHistoricalInfo(
			suite.ctx,
			height,
			&historicalInfo,
		)

		req := &stakingtypes.QueryHistoricalInfoRequest{
			Height: height,
		}
		res, err := suite.queryClient.HistoricalInfo(suite.ctx, req)
		suite.Require().NoError(err)

		suite.runHistoricalInfoIterations(req, res)
	})

	suite.SetupTest() // reset

	validator := suite.getStaticValidator()

	suite.stakingKeeper.SetValidator(suite.ctx, validator)
	suite.stakingKeeper.SetValidatorByPowerIndex(suite.ctx, validator)

	historicalInfo := stakingtypes.HistoricalInfo{
		Header: tmproto.Header{},
		Valset: []stakingtypes.Validator{validator},
	}

	height := int64(127)

	suite.stakingKeeper.SetHistoricalInfo(
		suite.ctx,
		height,
		&historicalInfo,
	)

	req := &stakingtypes.QueryHistoricalInfoRequest{
		Height: height,
	}
	res, err := suite.queryClient.HistoricalInfo(suite.ctx, req)
	suite.Require().NoError(err)

	suite.runHistoricalInfoIterations(req, res)
}

func (suite *DeterministicTestSuite) runDelegatorValidatorsIterations(req *stakingtypes.QueryDelegatorValidatorsRequest, prevDels *stakingtypes.QueryDelegatorValidatorsResponse) {
	for i := 0; i < 1000; i++ {
		res, err := suite.queryClient.DelegatorValidators(suite.ctx, req)
		suite.Require().NoError(err)

		suite.Require().Equal(res, prevDels)
	}
}

func (suite *DeterministicTestSuite) TestGRPCDelegatorValidators() {

	rapid.Check(suite.T(), func(t *rapid.T) {
		suite.SetupTest() // reset
		numVals := rapid.IntRange(1, 3).Draw(t, "num-dels")

		delegator := testdata.AddressGenerator(t).Draw(t, "delegator")

		for i := 0; i < numVals; i++ {
			amt := suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, rapid.Int64Range(100, 1000).Draw(t, "amount"))
			validator := suite.getValidatorWithStatus(t, stakingtypes.Bonded)
			suite.stakingKeeper.SetValidator(suite.ctx, validator)
			suite.stakingKeeper.SetValidatorByPowerIndex(suite.ctx, validator)

			// TODO remove mocks
			suite.bankKeeper.EXPECT().DelegateCoinsFromAccountToModule(
				suite.ctx, delegator, stakingtypes.BondedPoolName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, amt))).Return(nil)

			_, err := suite.stakingKeeper.Delegate(suite.ctx, delegator, amt, stakingtypes.Unbonded, validator, true)
			suite.Require().NoError(err)
		}

		req := &stakingtypes.QueryDelegatorValidatorsRequest{
			DelegatorAddr: delegator.String(),
			Pagination:    testdata.PaginationGenerator(t, uint64(numVals)).Draw(t, "pagination"),
		}

		res, err := suite.queryClient.DelegatorValidators(suite.ctx, req)
		suite.Require().NoError(err)
		suite.runDelegatorValidatorsIterations(req, res)
	})

	suite.SetupTest() // reset

	validator := suite.getStaticValidator()
	moduleName := stakingtypes.NotBondedPoolName

	if validator.GetStatus() == stakingtypes.Bonded {
		moduleName = stakingtypes.BondedPoolName
	}

	suite.stakingKeeper.SetValidator(suite.ctx, validator)
	suite.stakingKeeper.SetValidatorByPowerIndex(suite.ctx, validator)

	delegator1 := sdk.MustAccAddressFromBech32("cosmos1nph3cfzk6trsmfxkeu943nvach5qw4vwstnvkl")
	amt1 := suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, 101)

	suite.bankKeeper.EXPECT().DelegateCoinsFromAccountToModule(
		suite.ctx, delegator1, moduleName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, amt1))).Return(nil)

	_, err := suite.stakingKeeper.Delegate(suite.ctx, delegator1, amt1, stakingtypes.Unbonded, validator, true)
	suite.Require().NoError(err)

	req := &stakingtypes.QueryDelegatorValidatorsRequest{
		DelegatorAddr: delegator1.String(),
	}

	res, err := suite.queryClient.DelegatorValidators(suite.ctx, req)
	suite.Require().NoError(err)
	suite.runDelegatorValidatorsIterations(req, res)
}

// func (suite *DeterministicTestSuite) runPoolIterations(req *stakingtypes.QueryPoolRequest, prevDels *stakingtypes.QueryPoolResponse) {
// 	for i := 0; i < 1000; i++ {
// 		res, err := suite.queryClient.Pool(suite.ctx, req)
// 		suite.Require().NoError(err)

// 		suite.Require().Equal(res, prevDels)
// 	}
// }

// func (suite *DeterministicTestSuite) TestGRPCPool() {

// 	rapid.Check(suite.T(), func(t *rapid.T) {

// 		// amt := suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, rapid.Int64Range(100, 1000).Draw(t, "amount"))
// 		validator := suite.getValidator(t)
// 		suite.stakingKeeper.SetValidator(suite.ctx, validator)
// 		suite.stakingKeeper.SetValidatorByPowerIndex(suite.ctx, validator)

// 		// // TODO remove mocks
// 		// suite.bankKeeper.EXPECT().DelegateCoinsFromAccountToModule(
// 		// 	suite.ctx, delegator, stakingtypes.BondedPoolName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, amt))).Return(nil)

// 		// _, err := suite.stakingKeeper.Delegate(suite.ctx, delegator, amt, stakingtypes.Unbonded, validator, true)
// 		// suite.Require().NoError(err)

// 		req := &stakingtypes.QueryPoolRequest{}
// 		res, err := suite.queryClient.Pool(suite.ctx, req)
// 		suite.Require().NoError(err)
// 		suite.runPoolIterations(req, res)
// 	})

// 	suite.SetupTest() // reset

// 	validator := suite.getStaticValidator()

// 	suite.stakingKeeper.SetValidator(suite.ctx, validator)
// 	suite.stakingKeeper.SetValidatorByPowerIndex(suite.ctx, validator)

// 	req := &stakingtypes.QueryPoolRequest{}

// 	res, err := suite.queryClient.Pool(suite.ctx, req)
// 	suite.Require().NoError(err)
// 	suite.runPoolIterations(req, res)
// }

func (suite *DeterministicTestSuite) runRedelegationsIterations(req *stakingtypes.QueryRedelegationsRequest, prevDels *stakingtypes.QueryRedelegationsResponse) {
	for i := 0; i < 1000; i++ {
		res, err := suite.queryClient.Redelegations(suite.ctx, req)
		suite.Require().NoError(err)

		suite.Require().Equal(res, prevDels)
	}
}

func (suite *DeterministicTestSuite) TestGRPCValidatorDelegations() {

	rapid.Check(suite.T(), func(t *rapid.T) {
		suite.SetupTest() // reset
		validator := suite.getValidatorWithStatus(t, stakingtypes.Bonded)
		srcValAddr, err := sdk.ValAddressFromBech32(validator.OperatorAddress)
		suite.Require().NoError(err)
		suite.stakingKeeper.SetValidator(suite.ctx, validator)
		suite.stakingKeeper.SetValidatorByPowerIndex(suite.ctx, validator)

		validator2 := suite.getValidatorWithStatus(t, stakingtypes.Bonded)
		dstValAddr, err := sdk.ValAddressFromBech32(validator2.OperatorAddress)
		suite.Require().NoError(err)
		suite.stakingKeeper.SetValidator(suite.ctx, validator2)
		suite.stakingKeeper.SetValidatorByPowerIndex(suite.ctx, validator2)

		// numDels := rapid.IntRange(1, 5).Draw(t, "num-dels")
		moduleName := stakingtypes.NotBondedPoolName
		if validator.GetStatus() == stakingtypes.Bonded {
			moduleName = stakingtypes.BondedPoolName
		}

		// for i := 0; i < numDels; i++ {
		delegator := testdata.AddressGenerator(t).Draw(t, "delegator")
		amt := suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, rapid.Int64Range(100, 1000).Draw(t, "amount"))

		// TODO remove mocks
		suite.bankKeeper.EXPECT().DelegateCoinsFromAccountToModule(
			suite.ctx, delegator, moduleName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, amt))).Return(nil)

		shares, err := suite.stakingKeeper.Delegate(suite.ctx, delegator, amt, stakingtypes.Unbonded, validator, true)
		suite.Require().NoError(err)

		_, err = suite.stakingKeeper.BeginRedelegation(suite.ctx, delegator, srcValAddr, dstValAddr, shares)
		// }

		req := &stakingtypes.QueryRedelegationsRequest{
			SrcValidatorAddr: srcValAddr.String(),
			DstValidatorAddr: dstValAddr.String(),
			Pagination:       testdata.PaginationGenerator(t, uint64(numDels)).Draw(t, "pagination"),
		}

		res, err := suite.queryClient.ValidatorDelegations(suite.ctx, req)
		suite.Require().NoError(err)
		suite.runValidatorDelegationsIterations(req, res)
	})

	suite.SetupTest() // reset

	validator := suite.getStaticValidator()
	moduleName := stakingtypes.NotBondedPoolName

	if validator.GetStatus() == stakingtypes.Bonded {
		moduleName = stakingtypes.BondedPoolName
	}

	suite.stakingKeeper.SetValidator(suite.ctx, validator)
	suite.stakingKeeper.SetValidatorByPowerIndex(suite.ctx, validator)

	delegator1 := sdk.MustAccAddressFromBech32("cosmos1nph3cfzk6trsmfxkeu943nvach5qw4vwstnvkl")
	amt1 := suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, 101)

	suite.bankKeeper.EXPECT().DelegateCoinsFromAccountToModule(
		suite.ctx, delegator1, moduleName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, amt1))).Return(nil)

	_, err := suite.stakingKeeper.Delegate(suite.ctx, delegator1, amt1, stakingtypes.Unbonded, validator, true)
	suite.Require().NoError(err)

	delegator2 := sdk.MustAccAddressFromBech32("cosmos139f7kncmglres2nf3h4hc4tade85ekfr8sulz5")
	amt2 := suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, 102)
	suite.bankKeeper.EXPECT().DelegateCoinsFromAccountToModule(
		suite.ctx, delegator2, moduleName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, amt2))).Return(nil)
	_, err = suite.stakingKeeper.Delegate(suite.ctx, delegator2, amt2, stakingtypes.Unbonded, validator, true)
	suite.Require().NoError(err)

	req := &stakingtypes.QueryValidatorDelegationsRequest{
		ValidatorAddr: validator.OperatorAddress,
	}

	res, err := suite.queryClient.ValidatorDelegations(suite.ctx, req)
	suite.Require().NoError(err)
	suite.runValidatorDelegationsIterations(req, res)
}
