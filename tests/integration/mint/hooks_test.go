package mint

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	authkeeper "cosmossdk.io/x/auth/keeper"
	bankkeeper "cosmossdk.io/x/bank/keeper"
	"cosmossdk.io/x/mint/keeper"
	"cosmossdk.io/x/mint/types"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// Most values here are taken from mainnet genesis to mimic real-world behavior:
	// https://github.com/osmosis-labs/networks/raw/main/osmosis-1/genesis.json
	defaultGenesisEpochProvisions = "821917808219.178082191780821917"
	defaultEpochIdentifier        = "day"
	// actual value taken from mainnet for sanity checking calculations.
	defaultMainnetThirdenedProvisions                 = "547945205479.452055068493150684"
	defaultMintingRewardsDistributionStartEpoch int64 = 1
	defaultReductionPeriodInEpochs                    = 365
)

var defaultReductionFactor = math.LegacyNewDec(2).Quo(math.LegacyNewDec(3))

func TestAfterEpochEnd(t *testing.T) {
	var (
		maxArithmeticTolerance = math.LegacyNewDec(5)
		expectedSupply         = sdk.DefaultPowerReduction.ToLegacyDec()
	)

	defaultGenesisEpochProvisionsDec, err := math.LegacyNewDecFromStr(defaultGenesisEpochProvisions)
	require.NoError(t, err)

	defaultMainnetThirdenedProvisionsDec, err := math.LegacyNewDecFromStr(defaultMainnetThirdenedProvisions)
	require.NoError(t, err)

	testcases := map[string]struct {
		// Args.
		hookArgEpochNum int64

		// Presets.
		preExistingEpochNum     int64
		mintDenom               string
		epochIdentifier         string
		genesisEpochProvisions  math.LegacyDec
		reductionPeriodInEpochs int64
		reductionFactor         math.LegacyDec
		mintStartEpoch          int64

		// Expected results.
		expectedLastReductionEpochNum int64
		expectedDistribution          math.LegacyDec
		expectedError                 bool
	}{
		"before start epoch - no distributions": {
			hookArgEpochNum: defaultMintingRewardsDistributionStartEpoch - 1,

			mintDenom:               sdk.DefaultBondDenom,
			genesisEpochProvisions:  defaultGenesisEpochProvisionsDec,
			epochIdentifier:         defaultEpochIdentifier,
			reductionPeriodInEpochs: defaultReductionPeriodInEpochs,
			reductionFactor:         defaultReductionFactor,
			mintStartEpoch:          defaultMintingRewardsDistributionStartEpoch,

			expectedDistribution: math.LegacyZeroDec(),
		},
		"at start epoch - distributes": {
			hookArgEpochNum: defaultMintingRewardsDistributionStartEpoch,

			mintDenom:               sdk.DefaultBondDenom,
			genesisEpochProvisions:  defaultGenesisEpochProvisionsDec,
			epochIdentifier:         defaultEpochIdentifier,
			reductionPeriodInEpochs: defaultReductionPeriodInEpochs,
			reductionFactor:         defaultReductionFactor,
			mintStartEpoch:          defaultMintingRewardsDistributionStartEpoch,

			expectedDistribution:          defaultGenesisEpochProvisionsDec,
			expectedLastReductionEpochNum: defaultMintingRewardsDistributionStartEpoch,
		},
		"after start epoch - distributes": {
			hookArgEpochNum: defaultMintingRewardsDistributionStartEpoch + 5,

			preExistingEpochNum:     defaultMintingRewardsDistributionStartEpoch,
			mintDenom:               sdk.DefaultBondDenom,
			genesisEpochProvisions:  defaultGenesisEpochProvisionsDec,
			epochIdentifier:         defaultEpochIdentifier,
			reductionPeriodInEpochs: defaultReductionPeriodInEpochs,
			reductionFactor:         defaultReductionFactor,
			mintStartEpoch:          defaultMintingRewardsDistributionStartEpoch,

			expectedDistribution:          defaultGenesisEpochProvisionsDec,
			expectedLastReductionEpochNum: defaultMintingRewardsDistributionStartEpoch,
		},
		"before reduction epoch - distributes, no reduction": {
			hookArgEpochNum: defaultReductionPeriodInEpochs,

			preExistingEpochNum:     defaultMintingRewardsDistributionStartEpoch,
			mintDenom:               sdk.DefaultBondDenom,
			genesisEpochProvisions:  defaultGenesisEpochProvisionsDec,
			epochIdentifier:         defaultEpochIdentifier,
			reductionPeriodInEpochs: defaultReductionPeriodInEpochs,
			reductionFactor:         defaultReductionFactor,
			mintStartEpoch:          defaultMintingRewardsDistributionStartEpoch,

			expectedDistribution:          defaultGenesisEpochProvisionsDec,
			expectedLastReductionEpochNum: defaultMintingRewardsDistributionStartEpoch,
		},
		"at reduction epoch - distributes, reduction occurs": {
			preExistingEpochNum: defaultMintingRewardsDistributionStartEpoch,

			hookArgEpochNum: defaultMintingRewardsDistributionStartEpoch + defaultReductionPeriodInEpochs,

			mintDenom:               sdk.DefaultBondDenom,
			genesisEpochProvisions:  defaultGenesisEpochProvisionsDec,
			epochIdentifier:         defaultEpochIdentifier,
			reductionPeriodInEpochs: defaultReductionPeriodInEpochs,
			reductionFactor:         defaultReductionFactor,
			mintStartEpoch:          defaultMintingRewardsDistributionStartEpoch,

			expectedDistribution:          defaultMainnetThirdenedProvisionsDec,
			expectedLastReductionEpochNum: defaultMintingRewardsDistributionStartEpoch + defaultReductionPeriodInEpochs,
		},
		"after reduction epoch - distributes, with reduced amounts": {
			hookArgEpochNum: defaultMintingRewardsDistributionStartEpoch + defaultReductionPeriodInEpochs,

			mintDenom:               sdk.DefaultBondDenom,
			genesisEpochProvisions:  defaultGenesisEpochProvisionsDec,
			epochIdentifier:         defaultEpochIdentifier,
			reductionPeriodInEpochs: defaultReductionPeriodInEpochs,
			reductionFactor:         defaultReductionFactor,
			mintStartEpoch:          defaultMintingRewardsDistributionStartEpoch,

			expectedDistribution:          defaultMainnetThirdenedProvisionsDec,
			expectedLastReductionEpochNum: defaultMintingRewardsDistributionStartEpoch + defaultReductionPeriodInEpochs,
		},
		"start epoch == reduction epoch = curEpoch": {
			hookArgEpochNum: defaultReductionPeriodInEpochs,

			mintDenom:               sdk.DefaultBondDenom,
			genesisEpochProvisions:  defaultGenesisEpochProvisionsDec,
			epochIdentifier:         defaultEpochIdentifier,
			reductionPeriodInEpochs: defaultReductionPeriodInEpochs,
			reductionFactor:         defaultReductionFactor,
			mintStartEpoch:          defaultReductionPeriodInEpochs,

			expectedDistribution:          defaultGenesisEpochProvisionsDec,
			expectedLastReductionEpochNum: defaultReductionPeriodInEpochs,
		},
		"start epoch == curEpoch + 1 && reduction epoch == curEpoch": {
			hookArgEpochNum: defaultReductionPeriodInEpochs,

			mintDenom:               sdk.DefaultBondDenom,
			genesisEpochProvisions:  defaultGenesisEpochProvisionsDec,
			epochIdentifier:         defaultEpochIdentifier,
			reductionPeriodInEpochs: defaultReductionPeriodInEpochs,
			reductionFactor:         defaultReductionFactor,
			mintStartEpoch:          defaultReductionPeriodInEpochs - 1,

			expectedDistribution:          defaultMainnetThirdenedProvisionsDec,
			expectedLastReductionEpochNum: defaultReductionPeriodInEpochs,
		},
		"start epoch > reduction epoch": {
			hookArgEpochNum: defaultReductionPeriodInEpochs,

			mintDenom:               sdk.DefaultBondDenom,
			genesisEpochProvisions:  defaultGenesisEpochProvisionsDec,
			epochIdentifier:         defaultEpochIdentifier,
			reductionPeriodInEpochs: defaultReductionPeriodInEpochs,
			reductionFactor:         defaultReductionFactor,
			mintStartEpoch:          defaultReductionPeriodInEpochs + 1,

			expectedDistribution: math.LegacyZeroDec(),
		},
		"custom epochIdentifier, at start epoch": {
			hookArgEpochNum: defaultMintingRewardsDistributionStartEpoch,

			mintDenom:               sdk.DefaultBondDenom,
			genesisEpochProvisions:  defaultGenesisEpochProvisionsDec,
			epochIdentifier:         "week",
			reductionPeriodInEpochs: defaultReductionPeriodInEpochs,
			reductionFactor:         defaultReductionFactor,
			mintStartEpoch:          defaultMintingRewardsDistributionStartEpoch,

			expectedDistribution: math.LegacyZeroDec(),
		},
		"custom genesisEpochProvisions, at start epoch": {
			hookArgEpochNum: defaultMintingRewardsDistributionStartEpoch,

			mintDenom:               sdk.DefaultBondDenom,
			genesisEpochProvisions:  math.LegacyNewDec(1_000_000_000),
			epochIdentifier:         defaultEpochIdentifier,
			reductionPeriodInEpochs: defaultReductionPeriodInEpochs,
			reductionFactor:         defaultReductionFactor,
			mintStartEpoch:          defaultMintingRewardsDistributionStartEpoch,

			expectedDistribution:          defaultGenesisEpochProvisionsDec,
			expectedLastReductionEpochNum: defaultMintingRewardsDistributionStartEpoch,
		},
		"custom reduction factor, reduction epoch": {
			hookArgEpochNum: defaultMintingRewardsDistributionStartEpoch + defaultReductionPeriodInEpochs,

			mintDenom:               sdk.DefaultBondDenom,
			genesisEpochProvisions:  defaultGenesisEpochProvisionsDec,
			epochIdentifier:         defaultEpochIdentifier,
			reductionPeriodInEpochs: defaultReductionPeriodInEpochs,
			reductionFactor:         math.LegacyNewDec(43).Quo(math.LegacyNewDec(55)),
			mintStartEpoch:          defaultMintingRewardsDistributionStartEpoch,

			expectedDistribution:          defaultGenesisEpochProvisionsDec.Mul(math.LegacyNewDec(43)).Quo(math.LegacyNewDec(55)),
			expectedLastReductionEpochNum: defaultMintingRewardsDistributionStartEpoch + defaultReductionPeriodInEpochs,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			mintParams := types.Params{
				MintDenom:                            tc.mintDenom,
				GenesisEpochProvisions:               tc.genesisEpochProvisions,
				EpochIdentifier:                      tc.epochIdentifier,
				ReductionPeriodInEpochs:              tc.reductionPeriodInEpochs,
				ReductionFactor:                      tc.reductionFactor,
				MintingRewardsDistributionStartEpoch: tc.mintStartEpoch,
			}

			var (
				mintKeeper    keeper.Keeper
				accountKeeper authkeeper.AccountKeeper
				bankKeeper    bankkeeper.Keeper
			)
			app, err := simtestutil.SetupAtGenesis(
				depinject.Configs(
					AppConfig,
					depinject.Supply(log.NewNopLogger()),
				),
				&mintKeeper,
				&accountKeeper,
				&bankKeeper,
			)
			require.NoError(t, err)
			ctx := app.BaseApp.NewContext(false)

			// Pre-set parameters and minter.
			err = mintKeeper.Params.Set(ctx, mintParams)
			require.NoError(t, err)
			err = mintKeeper.LastReductionEpoch.Set(ctx, tc.preExistingEpochNum)
			require.NoError(t, err)
			err = mintKeeper.Minter.Set(ctx, types.Minter{
				EpochProvisions: defaultGenesisEpochProvisionsDec,
			})
			require.NoError(t, err)

			if tc.expectedError {
				require.Error(t, mintKeeper.AfterEpochEnd(ctx, defaultEpochIdentifier, tc.hookArgEpochNum))
			} else {
				require.NoError(t, mintKeeper.AfterEpochEnd(ctx, defaultEpochIdentifier, tc.hookArgEpochNum))
			}

			// If panics, the behavior is undefined.
			if tc.expectedError {
				return
			}

			// Validate supply.
			math.LegacyDecApproxEq(t, expectedSupply.Add(tc.expectedDistribution), bankKeeper.GetSupply(ctx, sdk.DefaultBondDenom).Amount.ToLegacyDec(), maxArithmeticTolerance)

			// Validate epoch provisions.
			lastReductionEpochNum, err := mintKeeper.LastReductionEpoch.Get(ctx)
			require.NoError(t, err)
			require.Equal(t, tc.expectedLastReductionEpochNum, lastReductionEpochNum)

			if !tc.expectedDistribution.IsZero() {
				// Validate distribution.
				minter, err := mintKeeper.Minter.Get(ctx)
				require.NoError(t, err)
				math.LegacyDecApproxEq(t, tc.expectedDistribution, minter.EpochProvisions, math.LegacyNewDecWithPrec(1, 6))
			}
		})
	}
}
