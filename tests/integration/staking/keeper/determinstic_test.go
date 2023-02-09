package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"gotest.tools/v3/assert"
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

type deterministicFixture struct {
	ctx           sdk.Context
	stakingKeeper *stakingkeeper.Keeper
	bankKeeper    bankkeeper.BaseKeeper
	accountKeeper authkeeper.AccountKeeper
	queryClient   stakingtypes.QueryClient
	amt1          math.Int
	amt2          math.Int
}

func initDeterministicFixture(t *testing.T) *deterministicFixture {
	f := &deterministicFixture{}

	var interfaceRegistry codectypes.InterfaceRegistry

	app, err := simtestutil.Setup(
		stakingtestutil.AppConfig,
		&f.bankKeeper,
		&f.accountKeeper,
		&f.stakingKeeper,
		&interfaceRegistry,
	)
	assert.NilError(t, err)

	f.ctx = app.BaseApp.NewContext(false, cmtproto.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(f.ctx, interfaceRegistry)
	stakingtypes.RegisterQueryServer(queryHelper, stakingkeeper.Querier{Keeper: f.stakingKeeper})
	f.queryClient = stakingtypes.NewQueryClient(queryHelper)

	f.amt1 = f.stakingKeeper.TokensFromConsensusPower(f.ctx, 101)
	f.amt2 = f.stakingKeeper.TokensFromConsensusPower(f.ctx, 102)

	return f
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
func createValidator(rt *rapid.T, f *deterministicFixture, t *testing.T) stakingtypes.Validator {
	pubkey := pubKeyGenerator().Draw(rt, "pubkey")
	pubkeyAny, err := codectypes.NewAnyWithValue(&pubkey)
	assert.NilError(t, err)
	return stakingtypes.Validator{
		OperatorAddress: sdk.ValAddress(testdata.AddressGenerator(rt).Draw(rt, "address")).String(),
		ConsensusPubkey: pubkeyAny,
		Jailed:          rapid.Bool().Draw(rt, "jailed"),
		Status:          bondTypeGenerator().Draw(rt, "bond-status"),
		Tokens:          sdk.NewInt(rapid.Int64Min(10000).Draw(rt, "tokens")),
		DelegatorShares: sdk.NewDecWithPrec(rapid.Int64Range(1, 100).Draw(rt, "commission"), 2),
		Description: stakingtypes.NewDescription(
			rapid.StringN(5, 250, 255).Draw(rt, "moniker"),
			rapid.StringN(5, 250, 255).Draw(rt, "identity"),
			rapid.StringN(5, 250, 255).Draw(rt, "website"),
			rapid.StringN(5, 250, 255).Draw(rt, "securityContact"),
			rapid.StringN(5, 250, 255).Draw(rt, "details"),
		),
		UnbondingHeight: rapid.Int64Min(1).Draw(rt, "unbonding-height"),
		UnbondingTime:   time.Now().Add(durationGenerator().Draw(rt, "duration")),
		Commission: stakingtypes.NewCommission(
			sdk.NewDecWithPrec(rapid.Int64Range(0, 100).Draw(rt, "rate"), 2),
			sdk.NewDecWithPrec(rapid.Int64Range(0, 100).Draw(rt, "max-rate"), 2),
			sdk.NewDecWithPrec(rapid.Int64Range(0, 100).Draw(rt, "max-change-rate"), 2),
		),
		MinSelfDelegation: sdk.NewInt(rapid.Int64Min(1).Draw(rt, "tokens")),
	}
}

// createAndSetValidatorWithStatus creates a validator with random values but with given status and sets to the state
func createAndSetValidatorWithStatus(rt *rapid.T, f *deterministicFixture, t *testing.T, status stakingtypes.BondStatus) stakingtypes.Validator {
	val := createValidator(rt, f, t)
	val.Status = status
	setValidator(f, t, val)
	return val
}

// createAndSetValidator creates a validator with random values and sets to the state
func createAndSetValidator(rt *rapid.T, f *deterministicFixture, t *testing.T) stakingtypes.Validator {
	val := createValidator(rt, f, t)
	setValidator(f, t, val)
	return val
}

func setValidator(f *deterministicFixture, t *testing.T, validator stakingtypes.Validator) {
	f.stakingKeeper.SetValidator(f.ctx, validator)
	f.stakingKeeper.SetValidatorByPowerIndex(f.ctx, validator)
	f.stakingKeeper.SetValidatorByConsAddr(f.ctx, validator)
	assert.NilError(t, f.stakingKeeper.Hooks().AfterValidatorCreated(f.ctx, validator.GetOperator()))

	delegatorAddress := sdk.AccAddress(validator.GetOperator())
	coins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, validator.BondedTokens()))
	banktestutil.FundAccount(f.bankKeeper, f.ctx, delegatorAddress, coins)

	_, err := f.stakingKeeper.Delegate(f.ctx, delegatorAddress, validator.BondedTokens(), stakingtypes.Unbonded, validator, true)
	assert.NilError(t, err)
}

// getStaticValidator creates a validator with hard-coded values and sets to the state.
func getStaticValidator(f *deterministicFixture, t *testing.T) stakingtypes.Validator {
	pubkey := ed25519.PubKey{Key: []byte{24, 179, 242, 2, 151, 3, 34, 6, 1, 11, 0, 194, 202, 201, 77, 1, 167, 40, 249, 115, 32, 97, 18, 1, 1, 127, 255, 103, 13, 1, 34, 1}}
	pubkeyAny, err := codectypes.NewAnyWithValue(&pubkey)
	assert.NilError(t, err)

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

	setValidator(f, t, validator)
	return validator
}

// getStaticValidator2 creates a validator with hard-coded values and sets to the state.
func getStaticValidator2(f *deterministicFixture, t *testing.T) stakingtypes.Validator {
	pubkey := ed25519.PubKey{Key: []byte{40, 249, 115, 32, 97, 18, 1, 1, 127, 255, 103, 13, 1, 34, 1, 24, 179, 242, 2, 151, 3, 34, 6, 1, 11, 0, 194, 202, 201, 77, 1, 167}}
	pubkeyAny, err := codectypes.NewAnyWithValue(&pubkey)
	assert.NilError(t, err)

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
	setValidator(f, t, validator)

	return validator
}

// createDelegationAndDelegate funds the delegator account with a random delegation in range 100-1000 and delegates.
func createDelegationAndDelegate(rt *rapid.T, f *deterministicFixture, t *testing.T, delegator sdk.AccAddress, validator stakingtypes.Validator) (newShares math.LegacyDec, err error) {
	amt := f.stakingKeeper.TokensFromConsensusPower(f.ctx, rapid.Int64Range(100, 1000).Draw(rt, "amount"))
	return fundAccountAndDelegate(f, t, delegator, validator, amt)
}

// fundAccountAndDelegate funds the delegator account with the specified delegation and delegates.
func fundAccountAndDelegate(f *deterministicFixture, t *testing.T, delegator sdk.AccAddress, validator stakingtypes.Validator, amt math.Int) (newShares math.LegacyDec, err error) {
	coins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, amt))

	assert.NilError(t, f.bankKeeper.MintCoins(f.ctx, minttypes.ModuleName, coins))
	banktestutil.FundAccount(f.bankKeeper, f.ctx, delegator, coins)

	shares, err := f.stakingKeeper.Delegate(f.ctx, delegator, amt, stakingtypes.Unbonded, validator, true)
	return shares, err
}

func TestGRPCValidator(t *testing.T) {
	t.Parallel()
	f := initDeterministicFixture(t)

	rapid.Check(t, func(rt *rapid.T) {
		val := createAndSetValidator(rt, f, t)
		req := &stakingtypes.QueryValidatorRequest{
			ValidatorAddr: val.OperatorAddress,
		}

		testdata.DeterministicIterations(f.ctx, t, req, f.queryClient.Validator, 0, true)
	})

	f = initDeterministicFixture(t) // reset
	val := getStaticValidator(f, t)
	req := &stakingtypes.QueryValidatorRequest{
		ValidatorAddr: val.OperatorAddress,
	}

	testdata.DeterministicIterations(f.ctx, t, req, f.queryClient.Validator, 1915, false)
}

func TestGRPCValidators(t *testing.T) {
	t.Parallel()
	f := initDeterministicFixture(t)

	validatorStatus := []string{stakingtypes.BondStatusBonded, stakingtypes.BondStatusUnbonded, stakingtypes.BondStatusUnbonding}
	rapid.Check(t, func(rt *rapid.T) {
		valsCount := rapid.IntRange(1, 3).Draw(rt, "num-validators")
		for i := 0; i < valsCount; i++ {
			createAndSetValidator(rt, f, t)
		}

		req := &stakingtypes.QueryValidatorsRequest{
			Status:     validatorStatus[rapid.IntRange(0, 2).Draw(rt, "status")],
			Pagination: testdata.PaginationGenerator(rt, uint64(valsCount)).Draw(rt, "pagination"),
		}

		testdata.DeterministicIterations(f.ctx, t, req, f.queryClient.Validators, 0, true)
	})

	f = initDeterministicFixture(t) // reset
	getStaticValidator(f, t)
	getStaticValidator2(f, t)

	testdata.DeterministicIterations(f.ctx, t, &stakingtypes.QueryValidatorsRequest{}, f.queryClient.Validators, 3525, false)
}

func TestGRPCValidatorDelegations(t *testing.T) {
	t.Parallel()
	f := initDeterministicFixture(t)

	rapid.Check(t, func(rt *rapid.T) {
		validator := createAndSetValidatorWithStatus(rt, f, t, stakingtypes.Bonded)
		numDels := rapid.IntRange(1, 5).Draw(rt, "num-dels")

		for i := 0; i < numDels; i++ {
			delegator := testdata.AddressGenerator(rt).Draw(rt, "delegator")
			_, err := createDelegationAndDelegate(rt, f, t, delegator, validator)
			assert.NilError(t, err)
		}

		req := &stakingtypes.QueryValidatorDelegationsRequest{
			ValidatorAddr: validator.OperatorAddress,
			Pagination:    testdata.PaginationGenerator(rt, uint64(numDels)).Draw(rt, "pagination"),
		}

		testdata.DeterministicIterations(f.ctx, t, req, f.queryClient.ValidatorDelegations, 0, true)
	})

	f = initDeterministicFixture(t) // reset

	validator := getStaticValidator(f, t)

	_, err := fundAccountAndDelegate(f, t, delegatorAddr1, validator, f.amt1)
	assert.NilError(t, err)

	_, err = fundAccountAndDelegate(f, t, delegatorAddr2, validator, f.amt2)
	assert.NilError(t, err)

	req := &stakingtypes.QueryValidatorDelegationsRequest{
		ValidatorAddr: validator.OperatorAddress,
	}

	testdata.DeterministicIterations(f.ctx, t, req, f.queryClient.ValidatorDelegations, 11985, false)
}

func TestGRPCValidatorUnbondingDelegations(t *testing.T) {
	t.Parallel()
	f := initDeterministicFixture(t)

	rapid.Check(t, func(rt *rapid.T) {
		validator := createAndSetValidatorWithStatus(rt, f, t, stakingtypes.Bonded)
		numDels := rapid.IntRange(1, 3).Draw(rt, "num-dels")

		for i := 0; i < numDels; i++ {
			delegator := testdata.AddressGenerator(rt).Draw(rt, "delegator")
			shares, err := createDelegationAndDelegate(rt, f, t, delegator, validator)
			assert.NilError(t, err)

			_, _, err = f.stakingKeeper.Undelegate(f.ctx, delegator, validator.GetOperator(), shares)
			assert.NilError(t, err)
		}

		req := &stakingtypes.QueryValidatorUnbondingDelegationsRequest{
			ValidatorAddr: validator.OperatorAddress,
			Pagination:    testdata.PaginationGenerator(rt, uint64(numDels)).Draw(rt, "pagination"),
		}

		testdata.DeterministicIterations(f.ctx, t, req, f.queryClient.ValidatorUnbondingDelegations, 0, true)
	})

	f = initDeterministicFixture(t) // reset

	validator := getStaticValidator(f, t)
	shares1, err := fundAccountAndDelegate(f, t, delegatorAddr1, validator, f.amt1)
	assert.NilError(t, err)

	_, _, err = f.stakingKeeper.Undelegate(f.ctx, delegatorAddr1, validatorAddr1, shares1)
	assert.NilError(t, err)

	shares2, err := fundAccountAndDelegate(f, t, delegatorAddr2, validator, f.amt2)
	assert.NilError(t, err)

	_, _, err = f.stakingKeeper.Undelegate(f.ctx, delegatorAddr2, validatorAddr1, shares2)
	assert.NilError(t, err)

	req := &stakingtypes.QueryValidatorUnbondingDelegationsRequest{
		ValidatorAddr: validator.OperatorAddress,
	}

	testdata.DeterministicIterations(f.ctx, t, req, f.queryClient.ValidatorUnbondingDelegations, 3719, false)
}

func TestGRPCDelegation(t *testing.T) {
	t.Parallel()
	f := initDeterministicFixture(t)

	rapid.Check(t, func(rt *rapid.T) {
		validator := createAndSetValidatorWithStatus(rt, f, t, stakingtypes.Bonded)
		delegator := testdata.AddressGenerator(rt).Draw(rt, "delegator")
		_, err := createDelegationAndDelegate(rt, f, t, delegator, validator)
		assert.NilError(t, err)

		req := &stakingtypes.QueryDelegationRequest{
			ValidatorAddr: validator.OperatorAddress,
			DelegatorAddr: delegator.String(),
		}

		testdata.DeterministicIterations(f.ctx, t, req, f.queryClient.Delegation, 0, true)
	})

	f = initDeterministicFixture(t) // reset

	validator := getStaticValidator(f, t)
	_, err := fundAccountAndDelegate(f, t, delegatorAddr1, validator, f.amt1)
	assert.NilError(t, err)

	req := &stakingtypes.QueryDelegationRequest{
		ValidatorAddr: validator.OperatorAddress,
		DelegatorAddr: delegator1,
	}

	testdata.DeterministicIterations(f.ctx, t, req, f.queryClient.Delegation, 4635, false)
}

func TestGRPCUnbondingDelegation(t *testing.T) {
	t.Parallel()
	f := initDeterministicFixture(t)

	rapid.Check(t, func(rt *rapid.T) {
		validator := createAndSetValidatorWithStatus(rt, f, t, stakingtypes.Bonded)
		delegator := testdata.AddressGenerator(rt).Draw(rt, "delegator")
		shares, err := createDelegationAndDelegate(rt, f, t, delegator, validator)
		assert.NilError(t, err)

		_, _, err = f.stakingKeeper.Undelegate(f.ctx, delegator, validator.GetOperator(), shares)
		assert.NilError(t, err)

		req := &stakingtypes.QueryUnbondingDelegationRequest{
			ValidatorAddr: validator.OperatorAddress,
			DelegatorAddr: delegator.String(),
		}

		testdata.DeterministicIterations(f.ctx, t, req, f.queryClient.UnbondingDelegation, 0, true)
	})

	f = initDeterministicFixture(t) // reset
	validator := getStaticValidator(f, t)

	shares1, err := fundAccountAndDelegate(f, t, delegatorAddr1, validator, f.amt1)
	assert.NilError(t, err)

	_, _, err = f.stakingKeeper.Undelegate(f.ctx, delegatorAddr1, validatorAddr1, shares1)
	assert.NilError(t, err)

	req := &stakingtypes.QueryUnbondingDelegationRequest{
		ValidatorAddr: validator.OperatorAddress,
		DelegatorAddr: delegator1,
	}

	testdata.DeterministicIterations(f.ctx, t, req, f.queryClient.UnbondingDelegation, 1621, false)
}

func TestGRPCDelegatorDelegations(t *testing.T) {
	t.Parallel()
	f := initDeterministicFixture(t)

	rapid.Check(t, func(rt *rapid.T) {
		numVals := rapid.IntRange(1, 3).Draw(rt, "num-dels")
		delegator := testdata.AddressGenerator(rt).Draw(rt, "delegator")

		for i := 0; i < numVals; i++ {
			validator := createAndSetValidatorWithStatus(rt, f, t, stakingtypes.Bonded)
			_, err := createDelegationAndDelegate(rt, f, t, delegator, validator)
			assert.NilError(t, err)
		}

		req := &stakingtypes.QueryDelegatorDelegationsRequest{
			DelegatorAddr: delegator.String(),
			Pagination:    testdata.PaginationGenerator(rt, uint64(numVals)).Draw(rt, "pagination"),
		}

		testdata.DeterministicIterations(f.ctx, t, req, f.queryClient.DelegatorDelegations, 0, true)
	})

	f = initDeterministicFixture(t) // reset

	validator := getStaticValidator(f, t)
	_, err := fundAccountAndDelegate(f, t, delegatorAddr1, validator, f.amt1)
	assert.NilError(t, err)

	req := &stakingtypes.QueryDelegatorDelegationsRequest{
		DelegatorAddr: delegator1,
	}

	testdata.DeterministicIterations(f.ctx, t, req, f.queryClient.DelegatorDelegations, 4238, false)
}

func TestGRPCDelegatorValidator(t *testing.T) {
	t.Parallel()
	f := initDeterministicFixture(t)

	rapid.Check(t, func(rt *rapid.T) {
		validator := createAndSetValidatorWithStatus(rt, f, t, stakingtypes.Bonded)

		delegator := testdata.AddressGenerator(rt).Draw(rt, "delegator")
		_, err := createDelegationAndDelegate(rt, f, t, delegator, validator)
		assert.NilError(t, err)

		req := &stakingtypes.QueryDelegatorValidatorRequest{
			DelegatorAddr: delegator.String(),
			ValidatorAddr: validator.OperatorAddress,
		}

		testdata.DeterministicIterations(f.ctx, t, req, f.queryClient.DelegatorValidator, 0, true)
	})

	f = initDeterministicFixture(t) // reset

	validator := getStaticValidator(f, t)
	_, err := fundAccountAndDelegate(f, t, delegatorAddr1, validator, f.amt1)

	assert.NilError(t, err)

	req := &stakingtypes.QueryDelegatorValidatorRequest{
		DelegatorAddr: delegator1,
		ValidatorAddr: validator.OperatorAddress,
	}

	testdata.DeterministicIterations(f.ctx, t, req, f.queryClient.DelegatorValidator, 3563, false)
}

func TestGRPCDelegatorUnbondingDelegations(t *testing.T) {
	t.Parallel()
	f := initDeterministicFixture(t)

	rapid.Check(t, func(rt *rapid.T) {
		numVals := rapid.IntRange(1, 5).Draw(rt, "num-vals")
		delegator := testdata.AddressGenerator(rt).Draw(rt, "delegator")

		for i := 0; i < numVals; i++ {
			validator := createAndSetValidatorWithStatus(rt, f, t, stakingtypes.Bonded)
			shares, err := createDelegationAndDelegate(rt, f, t, delegator, validator)
			assert.NilError(t, err)

			_, _, err = f.stakingKeeper.Undelegate(f.ctx, delegator, validator.GetOperator(), shares)
			assert.NilError(t, err)
		}

		req := &stakingtypes.QueryDelegatorUnbondingDelegationsRequest{
			DelegatorAddr: delegator.String(),
			Pagination:    testdata.PaginationGenerator(rt, uint64(numVals)).Draw(rt, "pagination"),
		}

		testdata.DeterministicIterations(f.ctx, t, req, f.queryClient.DelegatorUnbondingDelegations, 0, true)
	})

	f = initDeterministicFixture(t) // reset

	validator := getStaticValidator(f, t)
	shares1, err := fundAccountAndDelegate(f, t, delegatorAddr1, validator, f.amt1)
	assert.NilError(t, err)

	_, _, err = f.stakingKeeper.Undelegate(f.ctx, delegatorAddr1, validatorAddr1, shares1)
	assert.NilError(t, err)

	req := &stakingtypes.QueryDelegatorUnbondingDelegationsRequest{
		DelegatorAddr: delegator1,
	}

	testdata.DeterministicIterations(f.ctx, t, req, f.queryClient.DelegatorUnbondingDelegations, 1302, false)
}

func TestGRPCHistoricalInfo(t *testing.T) {
	t.Parallel()
	f := initDeterministicFixture(t)

	rapid.Check(t, func(rt *rapid.T) {
		numVals := rapid.IntRange(1, 5).Draw(rt, "num-vals")
		vals := make(stakingtypes.Validators, 0, numVals)
		for i := 0; i < numVals; i++ {
			validator := createAndSetValidatorWithStatus(rt, f, t, stakingtypes.Bonded)
			vals = append(vals, validator)
		}

		historicalInfo := stakingtypes.HistoricalInfo{
			Header: cmtproto.Header{},
			Valset: vals,
		}

		height := rapid.Int64Min(0).Draw(rt, "height")

		f.stakingKeeper.SetHistoricalInfo(
			f.ctx,
			height,
			&historicalInfo,
		)

		req := &stakingtypes.QueryHistoricalInfoRequest{
			Height: height,
		}

		testdata.DeterministicIterations(f.ctx, t, req, f.queryClient.HistoricalInfo, 0, true)
	})

	f = initDeterministicFixture(t) // reset

	validator := getStaticValidator(f, t)

	historicalInfo := stakingtypes.HistoricalInfo{
		Header: cmtproto.Header{},
		Valset: []stakingtypes.Validator{validator},
	}

	height := int64(127)

	f.stakingKeeper.SetHistoricalInfo(
		f.ctx,
		height,
		&historicalInfo,
	)

	req := &stakingtypes.QueryHistoricalInfoRequest{
		Height: height,
	}

	testdata.DeterministicIterations(f.ctx, t, req, f.queryClient.HistoricalInfo, 1930, false)
}

func TestGRPCDelegatorValidators(t *testing.T) {
	t.Parallel()
	f := initDeterministicFixture(t)

	rapid.Check(t, func(rt *rapid.T) {
		numVals := rapid.IntRange(1, 3).Draw(rt, "num-dels")
		delegator := testdata.AddressGenerator(rt).Draw(rt, "delegator")

		for i := 0; i < numVals; i++ {
			validator := createAndSetValidatorWithStatus(rt, f, t, stakingtypes.Bonded)
			_, err := createDelegationAndDelegate(rt, f, t, delegator, validator)
			assert.NilError(t, err)
		}

		req := &stakingtypes.QueryDelegatorValidatorsRequest{
			DelegatorAddr: delegator.String(),
			Pagination:    testdata.PaginationGenerator(rt, uint64(numVals)).Draw(rt, "pagination"),
		}

		testdata.DeterministicIterations(f.ctx, t, req, f.queryClient.DelegatorValidators, 0, true)
	})

	f = initDeterministicFixture(t) // reset

	validator := getStaticValidator(f, t)

	_, err := fundAccountAndDelegate(f, t, delegatorAddr1, validator, f.amt1)
	assert.NilError(t, err)

	req := &stakingtypes.QueryDelegatorValidatorsRequest{DelegatorAddr: delegator1}
	testdata.DeterministicIterations(f.ctx, t, req, f.queryClient.DelegatorValidators, 3166, false)
}

func TestGRPCPool(t *testing.T) {
	t.Parallel()
	f := initDeterministicFixture(t)

	rapid.Check(t, func(rt *rapid.T) {
		createAndSetValidator(rt, f, t)

		testdata.DeterministicIterations(f.ctx, t, &stakingtypes.QueryPoolRequest{}, f.queryClient.Pool, 0, true)
	})

	f = initDeterministicFixture(t) // reset
	getStaticValidator(f, t)
	testdata.DeterministicIterations(f.ctx, t, &stakingtypes.QueryPoolRequest{}, f.queryClient.Pool, 6185, false)
}

func TestGRPCRedelegations(t *testing.T) {
	t.Parallel()
	f := initDeterministicFixture(t)

	rapid.Check(t, func(rt *rapid.T) {
		validator := createAndSetValidatorWithStatus(rt, f, t, stakingtypes.Bonded)
		srcValAddr, err := sdk.ValAddressFromBech32(validator.OperatorAddress)
		assert.NilError(t, err)

		validator2 := createAndSetValidatorWithStatus(rt, f, t, stakingtypes.Bonded)
		dstValAddr, err := sdk.ValAddressFromBech32(validator2.OperatorAddress)
		assert.NilError(t, err)

		numDels := rapid.IntRange(1, 5).Draw(rt, "num-dels")

		delegator := testdata.AddressGenerator(rt).Draw(rt, "delegator")
		shares, err := createDelegationAndDelegate(rt, f, t, delegator, validator)
		assert.NilError(t, err)

		_, err = f.stakingKeeper.BeginRedelegation(f.ctx, delegator, srcValAddr, dstValAddr, shares)
		assert.NilError(t, err)

		var req *stakingtypes.QueryRedelegationsRequest

		reqType := rapid.IntRange(0, 2).Draw(rt, "req-type")
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

		req.Pagination = testdata.PaginationGenerator(rt, uint64(numDels)).Draw(rt, "pagination")
		testdata.DeterministicIterations(f.ctx, t, req, f.queryClient.Redelegations, 0, true)
	})

	f = initDeterministicFixture(t) // reset
	validator := getStaticValidator(f, t)
	_ = getStaticValidator2(f, t)

	shares, err := fundAccountAndDelegate(f, t, delegatorAddr1, validator, f.amt1)
	assert.NilError(t, err)

	_, err = f.stakingKeeper.BeginRedelegation(f.ctx, delegatorAddr1, validatorAddr1, validatorAddr2, shares)
	assert.NilError(t, err)

	req := &stakingtypes.QueryRedelegationsRequest{
		DelegatorAddr:    delegator1,
		SrcValidatorAddr: validator1,
		DstValidatorAddr: validator2,
	}

	testdata.DeterministicIterations(f.ctx, t, req, f.queryClient.Redelegations, 3920, false)
}

func TestGRPCParams(t *testing.T) {
	t.Parallel()
	f := initDeterministicFixture(t)

	rapid.Check(t, func(rt *rapid.T) {
		params := stakingtypes.Params{
			BondDenom:         rapid.StringMatching(sdk.DefaultCoinDenomRegex()).Draw(rt, "bond-denom"),
			UnbondingTime:     durationGenerator().Draw(rt, "duration"),
			MaxValidators:     rapid.Uint32Min(1).Draw(rt, "max-validators"),
			MaxEntries:        rapid.Uint32Min(1).Draw(rt, "max-entries"),
			HistoricalEntries: rapid.Uint32Min(1).Draw(rt, "historical-entries"),
			MinCommissionRate: sdk.NewDecWithPrec(rapid.Int64Range(0, 100).Draw(rt, "commission"), 2),
		}

		err := f.stakingKeeper.SetParams(f.ctx, params)
		assert.NilError(t, err)

		testdata.DeterministicIterations(f.ctx, t, &stakingtypes.QueryParamsRequest{}, f.queryClient.Params, 0, true)
	})

	params := stakingtypes.Params{
		BondDenom:         "denom",
		UnbondingTime:     time.Hour,
		MaxValidators:     85,
		MaxEntries:        5,
		HistoricalEntries: 5,
		MinCommissionRate: sdk.NewDecWithPrec(5, 2),
	}

	err := f.stakingKeeper.SetParams(f.ctx, params)
	assert.NilError(t, err)

	testdata.DeterministicIterations(f.ctx, t, &stakingtypes.QueryParamsRequest{}, f.queryClient.Params, 1114, false)
}
