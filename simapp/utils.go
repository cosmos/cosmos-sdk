//nolint
package simapp

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"time"

	cmn "github.com/tendermint/tendermint/libs/common"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/genaccounts"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/supply"
)

//---------------------------------------------------------------------
// Flags

// List of SimApp flags for the simulator
var (
	flagGenesisFileValue        string
	flagParamsFileValue         string
	flagExportParamsPathValue   string
	flagExportParamsHeightValue int
	flagExportStatePathValue    string
	flagExportStatsPathValue    string
	flagSeedValue               int64
	flagInitialBlockHeightValue int
	flagNumBlocksValue          int
	flagBlockSizeValue          int
	flagLeanValue               bool
	flagCommitValue             bool
	flagOnOperationValue        bool // TODO: Remove in favor of binary search for invariant violation
	flagAllInvariantsValue      bool

	flagEnabledValue     bool
	flagVerboseValue     bool
	flagPeriodValue      int
	flagGenesisTimeValue int64
)

// GetSimulatorFlags gets the values of all the available simulation flags
func GetSimulatorFlags() {

	// Config fields
	flag.StringVar(&flagGenesisFileValue, "Genesis", "", "custom simulation genesis file; cannot be used with params file")
	flag.StringVar(&flagParamsFileValue, "Params", "", "custom simulation params file which overrides any random params; cannot be used with genesis")
	flag.StringVar(&flagExportParamsPathValue, "ExportParamsPath", "", "custom file path to save the exported params JSON")
	flag.IntVar(&flagExportParamsHeightValue, "ExportParamsHeight", 0, "height to which export the randomly generated params")
	flag.StringVar(&flagExportStatePathValue, "ExportStatePath", "", "custom file path to save the exported app state JSON")
	flag.StringVar(&flagExportStatsPathValue, "ExportStatsPath", "", "custom file path to save the exported simulation statistics JSON")
	flag.Int64Var(&flagSeedValue, "Seed", 42, "simulation random seed")
	flag.IntVar(&flagInitialBlockHeightValue, "InitialBlockHeight", 1, "initial block to start the simulation")
	flag.IntVar(&flagNumBlocksValue, "NumBlocks", 500, "number of new blocks to simulate from the initial block height")
	flag.IntVar(&flagBlockSizeValue, "BlockSize", 200, "operations per block")
	flag.BoolVar(&flagLeanValue, "Lean", false, "lean simulation log output")
	flag.BoolVar(&flagCommitValue, "Commit", false, "have the simulation commit")
	flag.BoolVar(&flagOnOperationValue, "SimulateEveryOperation", false, "run slow invariants every operation")
	flag.BoolVar(&flagAllInvariantsValue, "PrintAllInvariants", false, "print all invariants if a broken invariant is found")

	// SimApp flags
	flag.BoolVar(&flagEnabledValue, "Enabled", false, "enable the simulation")
	flag.BoolVar(&flagVerboseValue, "Verbose", false, "verbose log output")
	flag.IntVar(&flagPeriodValue, "Period", 1, "run slow invariants only once every period assertions")
	flag.Int64Var(&flagGenesisTimeValue, "GenesisTime", 0, "override genesis UNIX time instead of using a random UNIX time")
}

// NewConfigFromFlags creates a simulation from the retrieved values of the flags
func NewConfigFromFlags() simulation.Config {
	return simulation.Config{
		GenesisFile:        flagGenesisFileValue,
		ParamsFile:         flagParamsFileValue,
		ExportParamsPath:   flagExportParamsPathValue,
		ExportParamsHeight: flagExportParamsHeightValue,
		ExportStatePath:    flagExportStatePathValue,
		ExportStatsPath:    flagExportStatsPathValue,
		Seed:               flagSeedValue,
		InitialBlockHeight: flagInitialBlockHeightValue,
		NumBlocks:          flagNumBlocksValue,
		BlockSize:          flagBlockSizeValue,
		Lean:               flagLeanValue,
		Commit:             flagCommitValue,
		OnOperation:        flagOnOperationValue,
		AllInvariants:      flagAllInvariantsValue,
	}
}

// GenAuthGenesisState generates a random GenesisState for auth
func GenAuthGenesisState(cdc *codec.Codec, r *rand.Rand, ap simulation.AppParams, genesisState map[string]json.RawMessage) {
	authGenesis := auth.NewGenesisState(
		auth.NewParams(
			func(r *rand.Rand) uint64 {
				var v uint64
				ap.GetOrGenerate(cdc, simulation.MaxMemoChars, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.MaxMemoChars](r).(uint64)
					})
				return v
			}(r),
			func(r *rand.Rand) uint64 {
				var v uint64
				ap.GetOrGenerate(cdc, simulation.TxSigLimit, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.TxSigLimit](r).(uint64)
					})
				return v
			}(r),
			func(r *rand.Rand) uint64 {
				var v uint64
				ap.GetOrGenerate(cdc, simulation.TxSizeCostPerByte, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.TxSizeCostPerByte](r).(uint64)
					})
				return v
			}(r),
			func(r *rand.Rand) uint64 {
				var v uint64
				ap.GetOrGenerate(cdc, simulation.SigVerifyCostED25519, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.SigVerifyCostED25519](r).(uint64)
					})
				return v
			}(r),
			func(r *rand.Rand) uint64 {
				var v uint64
				ap.GetOrGenerate(cdc, simulation.SigVerifyCostSECP256K1, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.SigVerifyCostSECP256K1](r).(uint64)
					})
				return v
			}(r),
		),
	)

	fmt.Printf("Selected randomly generated auth parameters:\n%s\n", codec.MustMarshalJSONIndent(cdc, authGenesis.Params))
	genesisState[auth.ModuleName] = cdc.MustMarshalJSON(authGenesis)
}

// GenBankGenesisState generates a random GenesisState for bank
func GenBankGenesisState(cdc *codec.Codec, r *rand.Rand, ap simulation.AppParams, genesisState map[string]json.RawMessage) {
	bankGenesis := bank.NewGenesisState(
		func(r *rand.Rand) bool {
			var v bool
			ap.GetOrGenerate(cdc, simulation.SendEnabled, &v, r,
				func(r *rand.Rand) {
					v = simulation.ModuleParamSimulator[simulation.SendEnabled](r).(bool)
				})
			return v
		}(r),
	)

	fmt.Printf("Selected randomly generated bank parameters:\n%s\n", codec.MustMarshalJSONIndent(cdc, bankGenesis))
	genesisState[bank.ModuleName] = cdc.MustMarshalJSON(bankGenesis)
}

// GenSupplyGenesisState generates a random GenesisState for supply
func GenSupplyGenesisState(cdc *codec.Codec, amount, numInitiallyBonded, numAccs int64, genesisState map[string]json.RawMessage) {
	totalSupply := sdk.NewInt(amount * (numAccs + numInitiallyBonded))
	supplyGenesis := supply.NewGenesisState(
		sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, totalSupply)),
	)

	fmt.Printf("Generated supply parameters:\n%s\n", codec.MustMarshalJSONIndent(cdc, supplyGenesis))
	genesisState[supply.ModuleName] = cdc.MustMarshalJSON(supplyGenesis)
}

// GenGenesisAccounts generates a random GenesisState for the genesis accounts
func GenGenesisAccounts(
	cdc *codec.Codec, r *rand.Rand, accs []simulation.Account,
	genesisTimestamp time.Time, amount, numInitiallyBonded int64,
	genesisState map[string]json.RawMessage,
) {

	var genesisAccounts []genaccounts.GenesisAccount

	// randomly generate some genesis accounts
	for i, acc := range accs {
		coins := sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(amount))}
		bacc := auth.NewBaseAccountWithAddress(acc.Address)
		bacc.SetCoins(coins)

		var gacc genaccounts.GenesisAccount

		// Only consider making a vesting account once the initial bonded validator
		// set is exhausted due to needing to track DelegatedVesting.
		if int64(i) > numInitiallyBonded && r.Intn(100) < 50 {
			var (
				vacc    auth.VestingAccount
				endTime int64
			)

			startTime := genesisTimestamp.Unix()

			// Allow for some vesting accounts to vest very quickly while others very slowly.
			if r.Intn(100) < 50 {
				endTime = int64(simulation.RandIntBetween(r, int(startTime), int(startTime+(60*60*24*30))))
			} else {
				endTime = int64(simulation.RandIntBetween(r, int(startTime), int(startTime+(60*60*12))))
			}

			if startTime == endTime {
				endTime++
			}

			if r.Intn(100) < 50 {
				vacc = auth.NewContinuousVestingAccount(&bacc, startTime, endTime)
			} else {
				vacc = auth.NewDelayedVestingAccount(&bacc, endTime)
			}

			var err error
			gacc, err = genaccounts.NewGenesisAccountI(vacc)
			if err != nil {
				panic(err)
			}
		} else {
			gacc = genaccounts.NewGenesisAccount(&bacc)
		}

		genesisAccounts = append(genesisAccounts, gacc)
	}

	genesisState[genaccounts.ModuleName] = cdc.MustMarshalJSON(genesisAccounts)
}

// GenGovGenesisState generates a random GenesisState for gov
func GenGovGenesisState(cdc *codec.Codec, r *rand.Rand, ap simulation.AppParams, genesisState map[string]json.RawMessage) {
	var vp time.Duration
	ap.GetOrGenerate(cdc, simulation.VotingParamsVotingPeriod, &vp, r,
		func(r *rand.Rand) {
			vp = simulation.ModuleParamSimulator[simulation.VotingParamsVotingPeriod](r).(time.Duration)
		})

	govGenesis := gov.NewGenesisState(
		uint64(r.Intn(100)),
		gov.NewDepositParams(
			func(r *rand.Rand) sdk.Coins {
				var v sdk.Coins
				ap.GetOrGenerate(cdc, simulation.DepositParamsMinDeposit, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.DepositParamsMinDeposit](r).(sdk.Coins)
					})
				return v
			}(r),
			vp,
		),
		gov.NewVotingParams(vp),
		gov.NewTallyParams(
			func(r *rand.Rand) sdk.Dec {
				var v sdk.Dec
				ap.GetOrGenerate(cdc, simulation.TallyParamsQuorum, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.TallyParamsQuorum](r).(sdk.Dec)
					})
				return v
			}(r),
			func(r *rand.Rand) sdk.Dec {
				var v sdk.Dec
				ap.GetOrGenerate(cdc, simulation.TallyParamsThreshold, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.TallyParamsThreshold](r).(sdk.Dec)
					})
				return v
			}(r),
			func(r *rand.Rand) sdk.Dec {
				var v sdk.Dec
				ap.GetOrGenerate(cdc, simulation.TallyParamsVeto, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.TallyParamsVeto](r).(sdk.Dec)
					})
				return v
			}(r),
		),
	)

	fmt.Printf("Selected randomly generated governance parameters:\n%s\n", codec.MustMarshalJSONIndent(cdc, govGenesis))
	genesisState[gov.ModuleName] = cdc.MustMarshalJSON(govGenesis)
}

// GenMintGenesisState generates a random GenesisState for mint
func GenMintGenesisState(cdc *codec.Codec, r *rand.Rand, ap simulation.AppParams, genesisState map[string]json.RawMessage) {
	mintGenesis := mint.NewGenesisState(
		mint.InitialMinter(
			func(r *rand.Rand) sdk.Dec {
				var v sdk.Dec
				ap.GetOrGenerate(cdc, simulation.Inflation, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.Inflation](r).(sdk.Dec)
					})
				return v
			}(r),
		),
		mint.NewParams(
			sdk.DefaultBondDenom,
			func(r *rand.Rand) sdk.Dec {
				var v sdk.Dec
				ap.GetOrGenerate(cdc, simulation.InflationRateChange, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.InflationRateChange](r).(sdk.Dec)
					})
				return v
			}(r),
			func(r *rand.Rand) sdk.Dec {
				var v sdk.Dec
				ap.GetOrGenerate(cdc, simulation.InflationMax, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.InflationMax](r).(sdk.Dec)
					})
				return v
			}(r),
			func(r *rand.Rand) sdk.Dec {
				var v sdk.Dec
				ap.GetOrGenerate(cdc, simulation.InflationMin, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.InflationMin](r).(sdk.Dec)
					})
				return v
			}(r),
			func(r *rand.Rand) sdk.Dec {
				var v sdk.Dec
				ap.GetOrGenerate(cdc, simulation.GoalBonded, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.GoalBonded](r).(sdk.Dec)
					})
				return v
			}(r),
			uint64(60*60*8766/5),
		),
	)

	fmt.Printf("Selected randomly generated minting parameters:\n%s\n", codec.MustMarshalJSONIndent(cdc, mintGenesis.Params))
	genesisState[mint.ModuleName] = cdc.MustMarshalJSON(mintGenesis)
}

// GenDistrGenesisState generates a random GenesisState for distribution
func GenDistrGenesisState(cdc *codec.Codec, r *rand.Rand, ap simulation.AppParams, genesisState map[string]json.RawMessage) {
	distrGenesis := distribution.GenesisState{
		FeePool: distribution.InitialFeePool(),
		CommunityTax: func(r *rand.Rand) sdk.Dec {
			var v sdk.Dec
			ap.GetOrGenerate(cdc, simulation.CommunityTax, &v, r,
				func(r *rand.Rand) {
					v = simulation.ModuleParamSimulator[simulation.CommunityTax](r).(sdk.Dec)
				})
			return v
		}(r),
		BaseProposerReward: func(r *rand.Rand) sdk.Dec {
			var v sdk.Dec
			ap.GetOrGenerate(cdc, simulation.BaseProposerReward, &v, r,
				func(r *rand.Rand) {
					v = simulation.ModuleParamSimulator[simulation.BaseProposerReward](r).(sdk.Dec)
				})
			return v
		}(r),
		BonusProposerReward: func(r *rand.Rand) sdk.Dec {
			var v sdk.Dec
			ap.GetOrGenerate(cdc, simulation.BonusProposerReward, &v, r,
				func(r *rand.Rand) {
					v = simulation.ModuleParamSimulator[simulation.BonusProposerReward](r).(sdk.Dec)
				})
			return v
		}(r),
	}

	fmt.Printf("Selected randomly generated distribution parameters:\n%s\n", codec.MustMarshalJSONIndent(cdc, distrGenesis))
	genesisState[distribution.ModuleName] = cdc.MustMarshalJSON(distrGenesis)
}

// GenSlashingGenesisState generates a random GenesisState for slashing
func GenSlashingGenesisState(
	cdc *codec.Codec, r *rand.Rand, stakingGen staking.GenesisState,
	ap simulation.AppParams, genesisState map[string]json.RawMessage,
) {
	slashingGenesis := slashing.NewGenesisState(
		slashing.NewParams(
			stakingGen.Params.UnbondingTime,
			func(r *rand.Rand) int64 {
				var v int64
				ap.GetOrGenerate(cdc, simulation.SignedBlocksWindow, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.SignedBlocksWindow](r).(int64)
					})
				return v
			}(r),
			func(r *rand.Rand) sdk.Dec {
				var v sdk.Dec
				ap.GetOrGenerate(cdc, simulation.MinSignedPerWindow, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.MinSignedPerWindow](r).(sdk.Dec)
					})
				return v
			}(r),
			func(r *rand.Rand) time.Duration {
				var v time.Duration
				ap.GetOrGenerate(cdc, simulation.DowntimeJailDuration, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.DowntimeJailDuration](r).(time.Duration)
					})
				return v
			}(r),
			func(r *rand.Rand) sdk.Dec {
				var v sdk.Dec
				ap.GetOrGenerate(cdc, simulation.SlashFractionDoubleSign, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.SlashFractionDoubleSign](r).(sdk.Dec)
					})
				return v
			}(r),
			func(r *rand.Rand) sdk.Dec {
				var v sdk.Dec
				ap.GetOrGenerate(cdc, simulation.SlashFractionDowntime, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.SlashFractionDowntime](r).(sdk.Dec)
					})
				return v
			}(r),
		),
		nil,
		nil,
	)

	fmt.Printf("Selected randomly generated slashing parameters:\n%s\n", codec.MustMarshalJSONIndent(cdc, slashingGenesis.Params))
	genesisState[slashing.ModuleName] = cdc.MustMarshalJSON(slashingGenesis)
}

// GenStakingGenesisState generates a random GenesisState for staking
func GenStakingGenesisState(
	cdc *codec.Codec, r *rand.Rand, accs []simulation.Account, amount, numAccs, numInitiallyBonded int64,
	ap simulation.AppParams, genesisState map[string]json.RawMessage,
) staking.GenesisState {

	stakingGenesis := staking.NewGenesisState(
		staking.NewParams(
			func(r *rand.Rand) time.Duration {
				var v time.Duration
				ap.GetOrGenerate(cdc, simulation.UnbondingTime, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.UnbondingTime](r).(time.Duration)
					})
				return v
			}(r),
			func(r *rand.Rand) uint16 {
				var v uint16
				ap.GetOrGenerate(cdc, simulation.MaxValidators, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.MaxValidators](r).(uint16)
					})
				return v
			}(r),
			7,
			sdk.DefaultBondDenom,
		),
		nil,
		nil,
	)

	var (
		validators  []staking.Validator
		delegations []staking.Delegation
	)

	valAddrs := make([]sdk.ValAddress, numInitiallyBonded)
	for i := 0; i < int(numInitiallyBonded); i++ {
		valAddr := sdk.ValAddress(accs[i].Address)
		valAddrs[i] = valAddr

		validator := staking.NewValidator(valAddr, accs[i].PubKey, staking.Description{})
		validator.Tokens = sdk.NewInt(amount)
		validator.DelegatorShares = sdk.NewDec(amount)
		delegation := staking.NewDelegation(accs[i].Address, valAddr, sdk.NewDec(amount))
		validators = append(validators, validator)
		delegations = append(delegations, delegation)
	}

	stakingGenesis.Validators = validators
	stakingGenesis.Delegations = delegations

	fmt.Printf("Selected randomly generated staking parameters:\n%s\n", codec.MustMarshalJSONIndent(cdc, stakingGenesis.Params))
	genesisState[staking.ModuleName] = cdc.MustMarshalJSON(stakingGenesis)

	return stakingGenesis
}

//---------------------------------------------------------------------
// Simulation Utils

// GetSimulationLog unmarshals the KVPair's Value to the corresponding type based on the
// each's module store key and the prefix bytes of the KVPair's key.
func GetSimulationLog(storeName string, sdr sdk.StoreDecoderRegistry, cdc *codec.Codec, kvAs, kvBs []cmn.KVPair) (log string) {
	for i := 0; i < len(kvAs); i++ {

		if len(kvAs[i].Value) == 0 && len(kvBs[i].Value) == 0 {
			// skip if the value doesn't have any bytes
			continue
		}

		decoder, ok := sdr[storeName]
		if ok {
			log += decoder(cdc, kvAs[i], kvBs[i])
		} else {
			log += fmt.Sprintf("store A %X => %X\nstore B %X => %X\n", kvAs[i].Key, kvAs[i].Value, kvBs[i].Key, kvBs[i].Value)
		}
	}

	return
}
