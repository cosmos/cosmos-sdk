//nolint
package simapp

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"time"

	"github.com/tendermint/tendermint/crypto/secp256k1"
	cmn "github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/baseapp"
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

// List of available flags for the simulator
var (
	genesisFile        string
	paramsFile         string
	exportParamsPath   string
	exportParamsHeight int
	exportStatePath    string
	exportStatsPath    string
	seed               int64
	initialBlockHeight int
	numBlocks          int
	blockSize          int
	enabled            bool
	verbose            bool
	lean               bool
	commit             bool
	period             int
	onOperation        bool // TODO Remove in favor of binary search for invariant violation
	allInvariants      bool
	genesisTime        int64
)

// NewSimAppUNSAFE is used for debugging purposes only.
//
// NOTE: to not use this function with non-test code
func NewSimAppUNSAFE(logger log.Logger, db dbm.DB, traceStore io.Writer, loadLatest bool,
	invCheckPeriod uint, baseAppOptions ...func(*baseapp.BaseApp),
) (gapp *SimApp, keyMain, keyStaking *sdk.KVStoreKey, stakingKeeper staking.Keeper) {

	gapp = NewSimApp(logger, db, traceStore, loadLatest, invCheckPeriod, baseAppOptions...)
	return gapp, gapp.keys[baseapp.MainStoreKey], gapp.keys[staking.StoreKey], gapp.StakingKeeper
}

// AppStateFromGenesisFileFn util function to generate the genesis AppState
// from a genesis.json file
func AppStateFromGenesisFileFn(
	r *rand.Rand, _ []simulation.Account, _ time.Time,
) (json.RawMessage, []simulation.Account, string) {

	var genesis tmtypes.GenesisDoc
	cdc := MakeCodec()

	bytes, err := ioutil.ReadFile(genesisFile)
	if err != nil {
		panic(err)
	}

	cdc.MustUnmarshalJSON(bytes, &genesis)

	var appState GenesisState
	cdc.MustUnmarshalJSON(genesis.AppState, &appState)

	accounts := genaccounts.GetGenesisStateFromAppState(cdc, appState)

	var newAccs []simulation.Account
	for _, acc := range accounts {
		// Pick a random private key, since we don't know the actual key
		// This should be fine as it's only used for mock Tendermint validators
		// and these keys are never actually used to sign by mock Tendermint.
		privkeySeed := make([]byte, 15)
		r.Read(privkeySeed)

		privKey := secp256k1.GenPrivKeySecp256k1(privkeySeed)
		newAccs = append(newAccs, simulation.Account{privKey, privKey.PubKey(), acc.Address})
	}

	return genesis.AppState, newAccs, genesis.ChainID
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
