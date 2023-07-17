package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/suite"
	"pgregory.net/rapid"

	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtestutil "github.com/cosmos/cosmos-sdk/x/staking/testutil"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var (
	validator1        = "cosmosvaloper1qqqryrs09ggeuqszqygqyqd2tgqmsqzewacjj7"
	validatorAddr1, _ = sdk.ValAddressFromBech32(validator1)
	validator2        = "cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj"
	validatorAddr2, _ = sdk.ValAddressFromBech32(validator2)
	delegator1        = "cosmos1nph3cfzk6trsmfxkeu943nvach5qw4vwstnvkl"
	delegatorAddr1    = sdk.MustAccAddressFromBech32(delegator1)
	delegator2        = "cosmos139f7kncmglres2nf3h4hc4tade85ekfr8sulz5"
	delegatorAddr2    = sdk.MustAccAddressFromBech32(delegator2)
)

type DeterministicTestSuite struct {
	suite.Suite

	ctx           sdk.Context
	stakingKeeper *stakingkeeper.Keeper
	bankKeeper    bankkeeper.BaseKeeper
	accountKeeper authkeeper.AccountKeeper
	queryClient   stakingtypes.QueryClient
	amt1          math.Int
	amt2          math.Int
}

func (s *DeterministicTestSuite) SetupTest() {
	var interfaceRegistry codectypes.InterfaceRegistry

	app, err := simtestutil.Setup(
		stakingtestutil.AppConfig,
		&s.bankKeeper,
		&s.accountKeeper,
		&s.stakingKeeper,
		&interfaceRegistry,
	)
	s.Require().NoError(err)

	// s.ctx = app.BaseApp.NewContext(false, tmproto.Header{Height: 1, Time: time.Now()}).WithGasMeter(sdk.NewInfiniteGasMeter())
	s.ctx = app.BaseApp.NewContext(false, tmproto.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(s.ctx, interfaceRegistry)
	stakingtypes.RegisterQueryServer(queryHelper, stakingkeeper.Querier{Keeper: s.stakingKeeper})
	s.queryClient = stakingtypes.NewQueryClient(queryHelper)

	s.amt1 = s.stakingKeeper.TokensFromConsensusPower(s.ctx, 101)
	s.amt2 = s.stakingKeeper.TokensFromConsensusPower(s.ctx, 102)
}

func TestDeterministicTestSuite(t *testing.T) {
	suite.Run(t, new(DeterministicTestSuite))
}

func durationGenerator() *rapid.Generator[time.Duration] {
	return rapid.Custom(func(t *rapid.T) time.Duration {
		now := time.Now()
		// range from current time to 365days.
		duration := rapid.Int64Range(now.Unix(), 365*24*60*60*now.Unix()).Draw(t, "time")
		return time.Duration(duration)
	})
}

func pubKeyGenerator() *rapid.Generator[ed25519.PubKey] {
	return rapid.Custom(func(t *rapid.T) ed25519.PubKey {
		pkBz := rapid.SliceOfN(rapid.Byte(), 32, 32).Draw(t, "hex")
		return ed25519.PubKey{Key: pkBz}
	})
}

func bondTypeGenerator() *rapid.Generator[stakingtypes.BondStatus] {
	bond_types := []stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Unbonded, stakingtypes.Unbonding}
	return rapid.Custom(func(t *rapid.T) stakingtypes.BondStatus {
		return bond_types[rapid.IntRange(0, 2).Draw(t, "range")]
	})
}

// createValidator creates a validator with random values.
func (suite *DeterministicTestSuite) createValidator(t *rapid.T) stakingtypes.Validator {
	pubkey := pubKeyGenerator().Draw(t, "pubkey")
	pubkeyAny, err := codectypes.NewAnyWithValue(&pubkey)
	suite.Require().NoError(err)
	return stakingtypes.Validator{
		OperatorAddress: sdk.ValAddress(testdata.AddressGenerator(t).Draw(t, "address")).String(),
		ConsensusPubkey: pubkeyAny,
		Jailed:          rapid.Bool().Draw(t, "jailed"),
		Status:          bondTypeGenerator().Draw(t, "bond-status"),
		Tokens:          sdk.NewInt(rapid.Int64Min(10000).Draw(t, "tokens")),
		DelegatorShares: sdk.NewDecWithPrec(rapid.Int64Range(1, 100).Draw(t, "commission"), 2),
		Description: stakingtypes.NewDescription(
			rapid.StringN(5, 250, 255).Draw(t, "moniker"),
			rapid.StringN(5, 250, 255).Draw(t, "identity"),
			rapid.StringN(5, 250, 255).Draw(t, "website"),
			rapid.StringN(5, 250, 255).Draw(t, "securityContact"),
			rapid.StringN(5, 250, 255).Draw(t, "details"),
		),
		UnbondingHeight: rapid.Int64Min(1).Draw(t, "unbonding-height"),
		UnbondingTime:   time.Now().Add(durationGenerator().Draw(t, "duration")),
		Commission: stakingtypes.NewCommission(
			sdk.NewDecWithPrec(rapid.Int64Range(0, 100).Draw(t, "rate"), 2),
			sdk.NewDecWithPrec(rapid.Int64Range(0, 100).Draw(t, "max-rate"), 2),
			sdk.NewDecWithPrec(rapid.Int64Range(0, 100).Draw(t, "max-change-rate"), 2),
		),
		MinSelfDelegation: sdk.NewInt(rapid.Int64Min(1).Draw(t, "tokens")),
	}
}

// createAndSetValidatorWithStatus creates a validator with random values but with given status and sets to the state
func (suite *DeterministicTestSuite) createAndSetValidatorWithStatus(t *rapid.T, status stakingtypes.BondStatus) stakingtypes.Validator {
	val := suite.createValidator(t)
	val.Status = status
	suite.setValidator(val)
	return val
}

// createAndSetValidator creates a validator with random values and sets to the state
func (suite *DeterministicTestSuite) createAndSetValidator(t *rapid.T) stakingtypes.Validator {
	val := suite.createValidator(t)
	suite.setValidator(val)
	return val
}

func (suite *DeterministicTestSuite) setValidator(validator stakingtypes.Validator) {
	suite.stakingKeeper.SetValidator(suite.ctx, validator)
	suite.stakingKeeper.SetValidatorByPowerIndex(suite.ctx, validator)
	suite.stakingKeeper.SetValidatorByConsAddr(suite.ctx, validator)
	suite.Require().NoError(suite.stakingKeeper.Hooks().AfterValidatorCreated(suite.ctx, validator.GetOperator()))

	delegatorAddress := sdk.AccAddress(validator.GetOperator())
	coins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, validator.BondedTokens()))
	banktestutil.FundAccount(suite.bankKeeper, suite.ctx, delegatorAddress, coins)

	_, err := suite.stakingKeeper.Delegate(suite.ctx, delegatorAddress, validator.BondedTokens(), stakingtypes.Unbonded, validator, true)
	suite.Require().NoError(err)
}

// getStaticValidator creates a validator with hard-coded values and sets to the state.
func (suite *DeterministicTestSuite) getStaticValidator() stakingtypes.Validator {
	pubkey := ed25519.PubKey{Key: []byte{24, 179, 242, 2, 151, 3, 34, 6, 1, 11, 0, 194, 202, 201, 77, 1, 167, 40, 249, 115, 32, 97, 18, 1, 1, 127, 255, 103, 13, 1, 34, 1}}
	pubkeyAny, err := codectypes.NewAnyWithValue(&pubkey)
	suite.Require().NoError(err)

	validator := stakingtypes.Validator{
		OperatorAddress: validator1,
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

	suite.setValidator(validator)
	return validator
}

// getStaticValidator2 creates a validator with hard-coded values and sets to the state.
func (suite *DeterministicTestSuite) getStaticValidator2() stakingtypes.Validator {
	pubkey := ed25519.PubKey{Key: []byte{40, 249, 115, 32, 97, 18, 1, 1, 127, 255, 103, 13, 1, 34, 1, 24, 179, 242, 2, 151, 3, 34, 6, 1, 11, 0, 194, 202, 201, 77, 1, 167}}
	pubkeyAny, err := codectypes.NewAnyWithValue(&pubkey)
	suite.Require().NoError(err)

	validator := stakingtypes.Validator{
		OperatorAddress: validator2,
		ConsensusPubkey: pubkeyAny,
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
	suite.setValidator(validator)

	return validator
}

// createDelegationAndDelegate funds the delegator account with a random delegation in range 100-1000 and delegates.
func (suite *DeterministicTestSuite) createDelegationAndDelegate(t *rapid.T, delegator sdk.AccAddress, validator stakingtypes.Validator) (newShares math.LegacyDec, err error) {
	amt := suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, rapid.Int64Range(100, 1000).Draw(t, "amount"))
	return suite.fundAccountAndDelegate(delegator, validator, amt)
}

// fundAccountAndDelegate funds the delegator account with the specified delegation and delegates.
func (suite *DeterministicTestSuite) fundAccountAndDelegate(delegator sdk.AccAddress, validator stakingtypes.Validator, amt math.Int) (newShares math.LegacyDec, err error) {
	coins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, amt))

	suite.Require().NoError(suite.bankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, coins))
	banktestutil.FundAccount(suite.bankKeeper, suite.ctx, delegator, coins)

	shares, err := suite.stakingKeeper.Delegate(suite.ctx, delegator, amt, stakingtypes.Unbonded, validator, true)
	return shares, err
}

func (suite *DeterministicTestSuite) TestGRPCValidator() {
	rapid.Check(suite.T(), func(t *rapid.T) {
		val := suite.createAndSetValidator(t)
		req := &stakingtypes.QueryValidatorRequest{
			ValidatorAddr: val.OperatorAddress,
		}

		testdata.DeterministicIterations(suite.ctx, suite.Require(), req, suite.queryClient.Validator, 0, true)
	})

	suite.SetupTest() // reset
	val := suite.getStaticValidator()
	req := &stakingtypes.QueryValidatorRequest{
		ValidatorAddr: val.OperatorAddress,
	}

	testdata.DeterministicIterations(suite.ctx, suite.Require(), req, suite.queryClient.Validator, 1933, false)
}

func (suite *DeterministicTestSuite) TestGRPCValidators() {
	validatorStatus := []string{stakingtypes.BondStatusBonded, stakingtypes.BondStatusUnbonded, stakingtypes.BondStatusUnbonding, ""}
	rapid.Check(suite.T(), func(t *rapid.T) {
		valsCount := rapid.IntRange(1, 3).Draw(t, "num-validators")
		for i := 0; i < valsCount; i++ {
			suite.createAndSetValidator(t)
		}

		req := &stakingtypes.QueryValidatorsRequest{
			Status:     validatorStatus[rapid.IntRange(0, 3).Draw(t, "status")],
			Pagination: testdata.PaginationGenerator(t, uint64(valsCount)).Draw(t, "pagination"),
		}

		testdata.DeterministicIterations(suite.ctx, suite.Require(), req, suite.queryClient.Validators, 0, true)
	})

	suite.SetupTest() // reset
	suite.getStaticValidator()
	suite.getStaticValidator2()

	testdata.DeterministicIterations(suite.ctx, suite.Require(), &stakingtypes.QueryValidatorsRequest{}, suite.queryClient.Validators, 3597, false)
}

func (suite *DeterministicTestSuite) TestGRPCValidatorDelegations() {
	rapid.Check(suite.T(), func(t *rapid.T) {
		validator := suite.createAndSetValidatorWithStatus(t, stakingtypes.Bonded)
		numDels := rapid.IntRange(1, 5).Draw(t, "num-dels")

		for i := 0; i < numDels; i++ {
			delegator := testdata.AddressGenerator(t).Draw(t, "delegator")
			_, err := suite.createDelegationAndDelegate(t, delegator, validator)
			suite.Require().NoError(err)
		}

		req := &stakingtypes.QueryValidatorDelegationsRequest{
			ValidatorAddr: validator.OperatorAddress,
			Pagination:    testdata.PaginationGenerator(t, uint64(numDels)).Draw(t, "pagination"),
		}

		testdata.DeterministicIterations(suite.ctx, suite.Require(), req, suite.queryClient.ValidatorDelegations, 0, true)
	})

	suite.SetupTest() // reset

	validator := suite.getStaticValidator()

	_, err := suite.fundAccountAndDelegate(delegatorAddr1, validator, suite.amt1)
	suite.Require().NoError(err)

	_, err = suite.fundAccountAndDelegate(delegatorAddr2, validator, suite.amt2)
	suite.Require().NoError(err)

	req := &stakingtypes.QueryValidatorDelegationsRequest{
		ValidatorAddr: validator.OperatorAddress,
	}

	testdata.DeterministicIterations(suite.ctx, suite.Require(), req, suite.queryClient.ValidatorDelegations, 12615, false)
}

func (suite *DeterministicTestSuite) TestGRPCValidatorUnbondingDelegations() {
	rapid.Check(suite.T(), func(t *rapid.T) {
		validator := suite.createAndSetValidatorWithStatus(t, stakingtypes.Bonded)
		numDels := rapid.IntRange(1, 3).Draw(t, "num-dels")

		for i := 0; i < numDels; i++ {
			delegator := testdata.AddressGenerator(t).Draw(t, "delegator")
			shares, err := suite.createDelegationAndDelegate(t, delegator, validator)
			suite.Require().NoError(err)

			_, err = suite.stakingKeeper.Undelegate(suite.ctx, delegator, validator.GetOperator(), shares)
			suite.Require().NoError(err)
		}

		req := &stakingtypes.QueryValidatorUnbondingDelegationsRequest{
			ValidatorAddr: validator.OperatorAddress,
			Pagination:    testdata.PaginationGenerator(t, uint64(numDels)).Draw(t, "pagination"),
		}

		testdata.DeterministicIterations(suite.ctx, suite.Require(), req, suite.queryClient.ValidatorUnbondingDelegations, 0, true)
	})

	suite.SetupTest() // reset

	validator := suite.getStaticValidator()
	shares1, err := suite.fundAccountAndDelegate(delegatorAddr1, validator, suite.amt1)
	suite.Require().NoError(err)

	_, err = suite.stakingKeeper.Undelegate(suite.ctx, delegatorAddr1, validatorAddr1, shares1)
	suite.Require().NoError(err)

	shares2, err := suite.fundAccountAndDelegate(delegatorAddr2, validator, suite.amt2)
	suite.Require().NoError(err)

	_, err = suite.stakingKeeper.Undelegate(suite.ctx, delegatorAddr2, validatorAddr1, shares2)
	suite.Require().NoError(err)

	req := &stakingtypes.QueryValidatorUnbondingDelegationsRequest{
		ValidatorAddr: validator.OperatorAddress,
	}

	testdata.DeterministicIterations(suite.ctx, suite.Require(), req, suite.queryClient.ValidatorUnbondingDelegations, 3755, false)
}

func (suite *DeterministicTestSuite) TestGRPCDelegation() {
	rapid.Check(suite.T(), func(t *rapid.T) {
		validator := suite.createAndSetValidatorWithStatus(t, stakingtypes.Bonded)
		delegator := testdata.AddressGenerator(t).Draw(t, "delegator")
		_, err := suite.createDelegationAndDelegate(t, delegator, validator)
		suite.Require().NoError(err)

		req := &stakingtypes.QueryDelegationRequest{
			ValidatorAddr: validator.OperatorAddress,
			DelegatorAddr: delegator.String(),
		}

		testdata.DeterministicIterations(suite.ctx, suite.Require(), req, suite.queryClient.Delegation, 0, true)
	})

	suite.SetupTest() // reset

	validator := suite.getStaticValidator()
	_, err := suite.fundAccountAndDelegate(delegatorAddr1, validator, suite.amt1)
	suite.Require().NoError(err)

	req := &stakingtypes.QueryDelegationRequest{
		ValidatorAddr: validator.OperatorAddress,
		DelegatorAddr: delegator1,
	}

	testdata.DeterministicIterations(suite.ctx, suite.Require(), req, suite.queryClient.Delegation, 4845, false)
}

func (suite *DeterministicTestSuite) TestGRPCUnbondingDelegation() {
	rapid.Check(suite.T(), func(t *rapid.T) {
		validator := suite.createAndSetValidatorWithStatus(t, stakingtypes.Bonded)
		delegator := testdata.AddressGenerator(t).Draw(t, "delegator")
		shares, err := suite.createDelegationAndDelegate(t, delegator, validator)
		suite.Require().NoError(err)

		_, err = suite.stakingKeeper.Undelegate(suite.ctx, delegator, validator.GetOperator(), shares)
		suite.Require().NoError(err)

		req := &stakingtypes.QueryUnbondingDelegationRequest{
			ValidatorAddr: validator.OperatorAddress,
			DelegatorAddr: delegator.String(),
		}

		testdata.DeterministicIterations(suite.ctx, suite.Require(), req, suite.queryClient.UnbondingDelegation, 0, true)
	})

	suite.SetupTest() // reset
	validator := suite.getStaticValidator()

	shares1, err := suite.fundAccountAndDelegate(delegatorAddr1, validator, suite.amt1)
	suite.Require().NoError(err)

	_, err = suite.stakingKeeper.Undelegate(suite.ctx, delegatorAddr1, validatorAddr1, shares1)
	suite.Require().NoError(err)

	req := &stakingtypes.QueryUnbondingDelegationRequest{
		ValidatorAddr: validator.OperatorAddress,
		DelegatorAddr: delegator1,
	}

	testdata.DeterministicIterations(suite.ctx, suite.Require(), req, suite.queryClient.UnbondingDelegation, 1639, false)
}

func (suite *DeterministicTestSuite) TestGRPCDelegatorDelegations() {
	rapid.Check(suite.T(), func(t *rapid.T) {
		numVals := rapid.IntRange(1, 3).Draw(t, "num-dels")
		delegator := testdata.AddressGenerator(t).Draw(t, "delegator")

		for i := 0; i < numVals; i++ {
			validator := suite.createAndSetValidatorWithStatus(t, stakingtypes.Bonded)
			_, err := suite.createDelegationAndDelegate(t, delegator, validator)
			suite.Require().NoError(err)
		}

		req := &stakingtypes.QueryDelegatorDelegationsRequest{
			DelegatorAddr: delegator.String(),
			Pagination:    testdata.PaginationGenerator(t, uint64(numVals)).Draw(t, "pagination"),
		}

		testdata.DeterministicIterations(suite.ctx, suite.Require(), req, suite.queryClient.DelegatorDelegations, 0, true)
	})

	suite.SetupTest() // reset

	validator := suite.getStaticValidator()
	_, err := suite.fundAccountAndDelegate(delegatorAddr1, validator, suite.amt1)
	suite.Require().NoError(err)

	req := &stakingtypes.QueryDelegatorDelegationsRequest{
		DelegatorAddr: delegator1,
	}

	testdata.DeterministicIterations(suite.ctx, suite.Require(), req, suite.queryClient.DelegatorDelegations, 4448, false)
}

func (suite *DeterministicTestSuite) TestGRPCDelegatorValidator() {
	rapid.Check(suite.T(), func(t *rapid.T) {
		validator := suite.createAndSetValidatorWithStatus(t, stakingtypes.Bonded)

		delegator := testdata.AddressGenerator(t).Draw(t, "delegator")
		_, err := suite.createDelegationAndDelegate(t, delegator, validator)
		suite.Require().NoError(err)

		req := &stakingtypes.QueryDelegatorValidatorRequest{
			DelegatorAddr: delegator.String(),
			ValidatorAddr: validator.OperatorAddress,
		}

		testdata.DeterministicIterations(suite.ctx, suite.Require(), req, suite.queryClient.DelegatorValidator, 0, true)
	})

	suite.SetupTest() // reset

	validator := suite.getStaticValidator()
	_, err := suite.fundAccountAndDelegate(delegatorAddr1, validator, suite.amt1)

	suite.Require().NoError(err)

	req := &stakingtypes.QueryDelegatorValidatorRequest{
		DelegatorAddr: delegator1,
		ValidatorAddr: validator.OperatorAddress,
	}

	testdata.DeterministicIterations(suite.ctx, suite.Require(), req, suite.queryClient.DelegatorValidator, 3581, false)
}

func (suite *DeterministicTestSuite) TestGRPCDelegatorUnbondingDelegations() {
	rapid.Check(suite.T(), func(t *rapid.T) {
		numVals := rapid.IntRange(1, 5).Draw(t, "num-vals")
		delegator := testdata.AddressGenerator(t).Draw(t, "delegator")

		for i := 0; i < numVals; i++ {
			validator := suite.createAndSetValidatorWithStatus(t, stakingtypes.Bonded)
			shares, err := suite.createDelegationAndDelegate(t, delegator, validator)
			suite.Require().NoError(err)

			_, err = suite.stakingKeeper.Undelegate(suite.ctx, delegator, validator.GetOperator(), shares)
			suite.Require().NoError(err)
		}

		req := &stakingtypes.QueryDelegatorUnbondingDelegationsRequest{
			DelegatorAddr: delegator.String(),
			Pagination:    testdata.PaginationGenerator(t, uint64(numVals)).Draw(t, "pagination"),
		}

		testdata.DeterministicIterations(suite.ctx, suite.Require(), req, suite.queryClient.DelegatorUnbondingDelegations, 0, true)
	})

	suite.SetupTest() // reset

	validator := suite.getStaticValidator()
	shares1, err := suite.fundAccountAndDelegate(delegatorAddr1, validator, suite.amt1)
	suite.Require().NoError(err)

	_, err = suite.stakingKeeper.Undelegate(suite.ctx, delegatorAddr1, validatorAddr1, shares1)
	suite.Require().NoError(err)

	req := &stakingtypes.QueryDelegatorUnbondingDelegationsRequest{
		DelegatorAddr: delegator1,
	}

	testdata.DeterministicIterations(suite.ctx, suite.Require(), req, suite.queryClient.DelegatorUnbondingDelegations, 1338, false)
}

func (suite *DeterministicTestSuite) TestGRPCHistoricalInfo() {
	rapid.Check(suite.T(), func(t *rapid.T) {
		numVals := rapid.IntRange(1, 5).Draw(t, "num-vals")
		vals := make(stakingtypes.Validators, 0, numVals)
		for i := 0; i < numVals; i++ {
			validator := suite.createAndSetValidatorWithStatus(t, stakingtypes.Bonded)
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

		testdata.DeterministicIterations(suite.ctx, suite.Require(), req, suite.queryClient.HistoricalInfo, 0, true)
	})

	suite.SetupTest() // reset

	validator := suite.getStaticValidator()

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

	testdata.DeterministicIterations(suite.ctx, suite.Require(), req, suite.queryClient.HistoricalInfo, 1948, false)
}

func (suite *DeterministicTestSuite) TestGRPCDelegatorValidators() {
	rapid.Check(suite.T(), func(t *rapid.T) {
		numVals := rapid.IntRange(1, 3).Draw(t, "num-dels")
		delegator := testdata.AddressGenerator(t).Draw(t, "delegator")

		for i := 0; i < numVals; i++ {
			validator := suite.createAndSetValidatorWithStatus(t, stakingtypes.Bonded)
			_, err := suite.createDelegationAndDelegate(t, delegator, validator)
			suite.Require().NoError(err)
		}

		req := &stakingtypes.QueryDelegatorValidatorsRequest{
			DelegatorAddr: delegator.String(),
			Pagination:    testdata.PaginationGenerator(t, uint64(numVals)).Draw(t, "pagination"),
		}

		testdata.DeterministicIterations(suite.ctx, suite.Require(), req, suite.queryClient.DelegatorValidators, 0, true)
	})

	suite.SetupTest() // reset

	validator := suite.getStaticValidator()

	_, err := suite.fundAccountAndDelegate(delegatorAddr1, validator, suite.amt1)
	suite.Require().NoError(err)

	req := &stakingtypes.QueryDelegatorValidatorsRequest{DelegatorAddr: delegator1}
	testdata.DeterministicIterations(suite.ctx, suite.Require(), req, suite.queryClient.DelegatorValidators, 3184, false)
}

func (suite *DeterministicTestSuite) TestGRPCPool() {
	rapid.Check(suite.T(), func(t *rapid.T) {
		suite.createAndSetValidator(t)

		testdata.DeterministicIterations(suite.ctx, suite.Require(), &stakingtypes.QueryPoolRequest{}, suite.queryClient.Pool, 0, true)
	})

	suite.SetupTest() // reset
	suite.getStaticValidator()
	testdata.DeterministicIterations(suite.ctx, suite.Require(), &stakingtypes.QueryPoolRequest{}, suite.queryClient.Pool, 6377, false)
}

func (suite *DeterministicTestSuite) TestGRPCRedelegations() {
	rapid.Check(suite.T(), func(t *rapid.T) {
		validator := suite.createAndSetValidatorWithStatus(t, stakingtypes.Bonded)
		srcValAddr, err := sdk.ValAddressFromBech32(validator.OperatorAddress)
		suite.Require().NoError(err)

		validator2 := suite.createAndSetValidatorWithStatus(t, stakingtypes.Bonded)
		dstValAddr, err := sdk.ValAddressFromBech32(validator2.OperatorAddress)
		suite.Require().NoError(err)

		numDels := rapid.IntRange(1, 5).Draw(t, "num-dels")

		delegator := testdata.AddressGenerator(t).Draw(t, "delegator")
		shares, err := suite.createDelegationAndDelegate(t, delegator, validator)
		suite.Require().NoError(err)

		_, err = suite.stakingKeeper.BeginRedelegation(suite.ctx, delegator, srcValAddr, dstValAddr, shares)
		suite.Require().NoError(err)

		var req *stakingtypes.QueryRedelegationsRequest

		reqType := rapid.IntRange(0, 2).Draw(t, "req-type")
		switch reqType {
		case 0: // queries redelegation with delegator, source and destination validators combination.
			req = &stakingtypes.QueryRedelegationsRequest{
				DelegatorAddr:    delegator.String(),
				SrcValidatorAddr: srcValAddr.String(),
				DstValidatorAddr: dstValAddr.String(),
			}
		case 1: // queries redelegations of source validator.
			req = &stakingtypes.QueryRedelegationsRequest{
				SrcValidatorAddr: srcValAddr.String(),
			}
		case 2: // queries all redelegations of a delegator.
			req = &stakingtypes.QueryRedelegationsRequest{
				DelegatorAddr: delegator.String(),
			}
		}

		req.Pagination = testdata.PaginationGenerator(t, uint64(numDels)).Draw(t, "pagination")
		testdata.DeterministicIterations(suite.ctx, suite.Require(), req, suite.queryClient.Redelegations, 0, true)
	})

	suite.SetupTest() // reset
	validator := suite.getStaticValidator()
	_ = suite.getStaticValidator2()

	shares, err := suite.fundAccountAndDelegate(delegatorAddr1, validator, suite.amt1)
	suite.Require().NoError(err)

	_, err = suite.stakingKeeper.BeginRedelegation(suite.ctx, delegatorAddr1, validatorAddr1, validatorAddr2, shares)
	suite.Require().NoError(err)

	req := &stakingtypes.QueryRedelegationsRequest{
		DelegatorAddr:    delegator1,
		SrcValidatorAddr: validator1,
		DstValidatorAddr: validator2,
	}

	testdata.DeterministicIterations(suite.ctx, suite.Require(), req, suite.queryClient.Redelegations, 3938, false)
}

func (suite *DeterministicTestSuite) TestGRPCParams() {
	rapid.Check(suite.T(), func(t *rapid.T) {
		params := stakingtypes.Params{
			BondDenom:                 rapid.StringMatching(sdk.DefaultCoinDenomRegex()).Draw(t, "bond-denom"),
			UnbondingTime:             durationGenerator().Draw(t, "duration"),
			MaxValidators:             rapid.Uint32Min(1).Draw(t, "max-validators"),
			MaxEntries:                rapid.Uint32Min(1).Draw(t, "max-entries"),
			HistoricalEntries:         rapid.Uint32Min(1).Draw(t, "historical-entries"),
			MinCommissionRate:         sdk.NewDecWithPrec(rapid.Int64Range(0, 100).Draw(t, "commission"), 2),
			ValidatorBondFactor:       stakingtypes.ValidatorBondCapDisabled,
			GlobalLiquidStakingCap:    sdk.NewDecWithPrec(1, 0),
			ValidatorLiquidStakingCap: sdk.NewDecWithPrec(1, 0),
		}

		err := suite.stakingKeeper.SetParams(suite.ctx, params)
		suite.Require().NoError(err)

		testdata.DeterministicIterations(suite.ctx, suite.Require(), &stakingtypes.QueryParamsRequest{}, suite.queryClient.Params, 0, true)
	})

	params := stakingtypes.Params{
		BondDenom:                 "denom",
		UnbondingTime:             time.Hour,
		MaxValidators:             85,
		MaxEntries:                5,
		HistoricalEntries:         5,
		MinCommissionRate:         sdk.NewDecWithPrec(5, 2),
		ValidatorBondFactor:       stakingtypes.ValidatorBondCapDisabled,
		GlobalLiquidStakingCap:    sdk.NewDecWithPrec(1, 0),
		ValidatorLiquidStakingCap: sdk.NewDecWithPrec(1, 0),
	}

	err := suite.stakingKeeper.SetParams(suite.ctx, params)
	suite.Require().NoError(err)

	testdata.DeterministicIterations(suite.ctx, suite.Require(), &stakingtypes.QueryParamsRequest{}, suite.queryClient.Params, 1306, false)
}
