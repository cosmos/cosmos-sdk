package simulation

import (
	"math/rand"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// Minimum time per block
	minTimePerBlock int64 = 10000 / 2

	// Maximum time per block
	maxTimePerBlock int64 = 10000
)

// TODO explain transitional matrix usage
var (
	// Currently there are 3 different liveness types,
	// fully online, spotty connection, offline.
	defaultLivenessTransitionMatrix, _ = CreateTransitionMatrix([][]int{
		{90, 20, 1},
		{10, 50, 5},
		{0, 10, 1000},
	})

	// 3 states: rand in range [0, 4*provided blocksize],
	// rand in range [0, 2 * provided blocksize], 0
	defaultBlockSizeTransitionMatrix, _ = CreateTransitionMatrix([][]int{
		{85, 5, 0},
		{15, 92, 1},
		{0, 3, 99},
	})

	// ModuleParamSimulator defines module parameter value simulators. All
	// values simulated should be within valid acceptable range for the given
	// parameter.
	ModuleParamSimulator = map[string]func(r *rand.Rand) interface{}{
		"MaxMemoCharacters": func(r *rand.Rand) interface{} {
			return uint64(RandIntBetween(r, 100, 200))
		},
		"TxSigLimit": func(r *rand.Rand) interface{} {
			return uint64(r.Intn(7) + 1)
		},
		"TxSizeCostPerByte": func(r *rand.Rand) interface{} {
			return uint64(RandIntBetween(r, 5, 15))
		},
		"SigVerifyCostED25519": func(r *rand.Rand) interface{} {
			return uint64(RandIntBetween(r, 500, 1000))
		},
		"SigVerifyCostSecp256k1": func(r *rand.Rand) interface{} {
			return uint64(RandIntBetween(r, 500, 1000))
		},
		"DepositParams/MinDeposit": func(r *rand.Rand) interface{} {
			return sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, int64(RandIntBetween(r, 1, 1e3)))}
		},
		"VotingParams/VotingPeriod": func(r *rand.Rand) interface{} {
			return time.Duration(RandIntBetween(r, 1, 2*60*60*24*2)) * time.Second
		},
		"TallyParams/Quorum": func(r *rand.Rand) interface{} {
			return sdk.NewDecWithPrec(int64(RandIntBetween(r, 334, 500)), 3)
		},
		"TallyParams/Threshold": func(r *rand.Rand) interface{} {
			return sdk.NewDecWithPrec(int64(RandIntBetween(r, 450, 550)), 3)
		},
		"TallyParams/Veto": func(r *rand.Rand) interface{} {
			return sdk.NewDecWithPrec(int64(RandIntBetween(r, 250, 334)), 3)
		},
		"UnbondingTime": func(r *rand.Rand) interface{} {
			return time.Duration(RandIntBetween(r, 60, 60*60*24*3*2)) * time.Second
		},
		"MaxValidators": func(r *rand.Rand) interface{} {
			return uint16(r.Intn(250) + 1)
		},
		"SignedBlocksWindow": func(r *rand.Rand) interface{} {
			return int64(RandIntBetween(r, 10, 1000))
		},
		"MinSignedPerWindow": func(r *rand.Rand) interface{} {
			return sdk.NewDecWithPrec(int64(r.Intn(10)), 1)
		},
		"DowntimeJailDuration": func(r *rand.Rand) interface{} {
			return time.Duration(RandIntBetween(r, 60, 60*60*24)) * time.Second
		},
		"SlashFractionDoubleSign": func(r *rand.Rand) interface{} {
			return sdk.NewDec(1).Quo(sdk.NewDec(int64(r.Intn(50) + 1)))
		},
		"SlashFractionDowntime": func(r *rand.Rand) interface{} {
			return sdk.NewDec(1).Quo(sdk.NewDec(int64(r.Intn(200) + 1)))
		},
		"InflationRateChange": func(r *rand.Rand) interface{} {
			return sdk.NewDecWithPrec(int64(r.Intn(99)), 2)
		},
		"InflationMax": func(r *rand.Rand) interface{} {
			return sdk.NewDecWithPrec(20, 2)
		},
		"InflationMin": func(r *rand.Rand) interface{} {
			return sdk.NewDecWithPrec(7, 2)
		},
		"GoalBonded": func(r *rand.Rand) interface{} {
			return sdk.NewDecWithPrec(67, 2)
		},
	}
)

// Simulation parameters
type Params struct {
	PastEvidenceFraction      float64
	NumKeys                   int
	EvidenceFraction          float64
	InitialLivenessWeightings []int
	LivenessTransitionMatrix  TransitionMatrix
	BlockSizeTransitionMatrix TransitionMatrix
}

// Return default simulation parameters
func DefaultParams() Params {
	return Params{
		PastEvidenceFraction:      0.5,
		NumKeys:                   250,
		EvidenceFraction:          0.5,
		InitialLivenessWeightings: []int{40, 5, 5},
		LivenessTransitionMatrix:  defaultLivenessTransitionMatrix,
		BlockSizeTransitionMatrix: defaultBlockSizeTransitionMatrix,
	}
}

// Return random simulation parameters
func RandomParams(r *rand.Rand) Params {
	return Params{
		PastEvidenceFraction:      r.Float64(),
		NumKeys:                   RandIntBetween(r, 2, 250),
		EvidenceFraction:          r.Float64(),
		InitialLivenessWeightings: []int{RandIntBetween(r, 1, 80), r.Intn(10), r.Intn(10)},
		LivenessTransitionMatrix:  defaultLivenessTransitionMatrix,
		BlockSizeTransitionMatrix: defaultBlockSizeTransitionMatrix,
	}
}
