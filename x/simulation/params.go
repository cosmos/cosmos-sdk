package simulation

import (
	"encoding/json"
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// nolint
const (
	// Minimum time per block
	minTimePerBlock int64 = 10000 / 2

	// Maximum time per block
	maxTimePerBlock int64 = 10000

	// Simulation parameter constants
	SendEnabled              = "send_enabled"
	MaxMemoChars             = "max_memo_characters"
	TxSigLimit               = "tx_sig_limit"
	TxSizeCostPerByte        = "tx_size_cost_per_byte"
	SigVerifyCostED25519     = "sig_verify_cost_ed25519"
	SigVerifyCostSECP256K1   = "sig_verify_cost_secp256k1"
	DepositParamsMinDeposit  = "deposit_params_min_deposit"
	VotingParamsVotingPeriod = "voting_params_voting_period"
	TallyParamsQuorum        = "tally_params_quorum"
	TallyParamsThreshold     = "tally_params_threshold"
	TallyParamsVeto          = "tally_params_veto"
	UnbondingTime            = "unbonding_time"
	MaxValidators            = "max_validators"
	SignedBlocksWindow       = "signed_blocks_window"
	MinSignedPerWindow       = "min_signed_per_window"
	DowntimeJailDuration     = "downtime_jail_duration"
	SlashFractionDoubleSign  = "slash_fraction_double_sign"
	SlashFractionDowntime    = "slash_fraction_downtime"
	InflationRateChange      = "inflation_rate_change"
	Inflation                = "inflation"
	InflationMax             = "inflation_max"
	InflationMin             = "inflation_min"
	GoalBonded               = "goal_bonded"
	CommunityTax             = "community_tax"
	BaseProposerReward       = "base_proposer_reward"
	BonusProposerReward      = "bonus_proposer_reward"
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
		SendEnabled: func(r *rand.Rand) interface{} {
			return r.Int63n(2) == 0
		},
		MaxMemoChars: func(r *rand.Rand) interface{} {
			return uint64(RandIntBetween(r, 100, 200))
		},
		TxSigLimit: func(r *rand.Rand) interface{} {
			return uint64(r.Intn(7) + 1)
		},
		TxSizeCostPerByte: func(r *rand.Rand) interface{} {
			return uint64(RandIntBetween(r, 5, 15))
		},
		SigVerifyCostED25519: func(r *rand.Rand) interface{} {
			return uint64(RandIntBetween(r, 500, 1000))
		},
		SigVerifyCostSECP256K1: func(r *rand.Rand) interface{} {
			return uint64(RandIntBetween(r, 500, 1000))
		},
		DepositParamsMinDeposit: func(r *rand.Rand) interface{} {
			return sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, int64(RandIntBetween(r, 1, 1e3)))}
		},
		VotingParamsVotingPeriod: func(r *rand.Rand) interface{} {
			return time.Duration(RandIntBetween(r, 1, 2*60*60*24*2)) * time.Second
		},
		TallyParamsQuorum: func(r *rand.Rand) interface{} {
			return sdk.NewDecWithPrec(int64(RandIntBetween(r, 334, 500)), 3)
		},
		TallyParamsThreshold: func(r *rand.Rand) interface{} {
			return sdk.NewDecWithPrec(int64(RandIntBetween(r, 450, 550)), 3)
		},
		TallyParamsVeto: func(r *rand.Rand) interface{} {
			return sdk.NewDecWithPrec(int64(RandIntBetween(r, 250, 334)), 3)
		},
		UnbondingTime: func(r *rand.Rand) interface{} {
			return time.Duration(RandIntBetween(r, 60, 60*60*24*3*2)) * time.Second
		},
		MaxValidators: func(r *rand.Rand) interface{} {
			return uint16(r.Intn(250) + 1)
		},
		SignedBlocksWindow: func(r *rand.Rand) interface{} {
			return int64(RandIntBetween(r, 10, 1000))
		},
		MinSignedPerWindow: func(r *rand.Rand) interface{} {
			return sdk.NewDecWithPrec(int64(r.Intn(10)), 1)
		},
		DowntimeJailDuration: func(r *rand.Rand) interface{} {
			return time.Duration(RandIntBetween(r, 60, 60*60*24)) * time.Second
		},
		SlashFractionDoubleSign: func(r *rand.Rand) interface{} {
			return sdk.NewDec(1).Quo(sdk.NewDec(int64(r.Intn(50) + 1)))
		},
		SlashFractionDowntime: func(r *rand.Rand) interface{} {
			return sdk.NewDec(1).Quo(sdk.NewDec(int64(r.Intn(200) + 1)))
		},
		InflationRateChange: func(r *rand.Rand) interface{} {
			return sdk.NewDecWithPrec(int64(r.Intn(99)), 2)
		},
		Inflation: func(r *rand.Rand) interface{} {
			return sdk.NewDecWithPrec(int64(r.Intn(99)), 2)
		},
		InflationMax: func(r *rand.Rand) interface{} {
			return sdk.NewDecWithPrec(20, 2)
		},
		InflationMin: func(r *rand.Rand) interface{} {
			return sdk.NewDecWithPrec(7, 2)
		},
		GoalBonded: func(r *rand.Rand) interface{} {
			return sdk.NewDecWithPrec(67, 2)
		},
		CommunityTax: func(r *rand.Rand) interface{} {
			return sdk.NewDecWithPrec(1, 2).Add(sdk.NewDecWithPrec(int64(r.Intn(30)), 2))
		},
		BaseProposerReward: func(r *rand.Rand) interface{} {
			return sdk.NewDecWithPrec(1, 2).Add(sdk.NewDecWithPrec(int64(r.Intn(30)), 2))
		},
		BonusProposerReward: func(r *rand.Rand) interface{} {
			return sdk.NewDecWithPrec(1, 2).Add(sdk.NewDecWithPrec(int64(r.Intn(30)), 2))
		},
	}
)

// TODO add description
type (
	AppParams      map[string]json.RawMessage
	ParamSimulator func(r *rand.Rand)
)

// GetOrGenerate attempts to get a given parameter by key from the AppParams
// object. If it exists, it'll be decoded and returned. Otherwise, the provided
// ParamSimulator is used to generate a random value.
func (sp AppParams) GetOrGenerate(cdc *codec.Codec, key string, ptr interface{}, r *rand.Rand, ps ParamSimulator) {
	if v, ok := sp[key]; ok && v != nil {
		cdc.MustUnmarshalJSON(v, ptr)
		return
	}

	ps(r)
}

// Simulation parameters
type Params struct {
	PastEvidenceFraction      float64
	NumKeys                   int
	EvidenceFraction          float64
	InitialLivenessWeightings []int
	LivenessTransitionMatrix  TransitionMatrix
	BlockSizeTransitionMatrix TransitionMatrix
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
