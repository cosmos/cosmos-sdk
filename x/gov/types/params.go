package types

import (
	"fmt"
	"time"

	yaml "gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Default period for deposits & voting
const (
	DefaultPeriod time.Duration = time.Hour * 24 * 2 // 2 days
)

// Default governance params
var (
	DefaultMinDepositTokens = sdk.TokensFromConsensusPower(10)
	DefaultQuorum           = sdk.NewDecWithPrec(334, 3)
	DefaultThreshold        = sdk.NewDecWithPrec(5, 1)
	DefaultVetoThreshold    = sdk.NewDecWithPrec(334, 3)
)

// Parameter store key
var (
	ParamStoreKeyDepositParams = []byte("depositparams")
	ParamStoreKeyVotingParams  = []byte("votingparams")
	ParamStoreKeyTallyParams   = []byte("tallyparams")
)

// ParamKeyTable - Key declaration for parameters
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable(
		paramtypes.NewParamSetPair(ParamStoreKeyDepositParams, DepositParams{}, validateDepositParams),
		paramtypes.NewParamSetPair(ParamStoreKeyVotingParams, VotingParams{}, validateVotingParams),
		paramtypes.NewParamSetPair(ParamStoreKeyTallyParams, TallyParams{}, validateTallyParams),
	)
}

// NewDepositParams creates a new DepositParams object
func NewDepositParams(minDeposit sdk.Coins, maxDepositPeriod time.Duration) DepositParams {
	return DepositParams{
		MinDeposit:       minDeposit,
		MaxDepositPeriod: maxDepositPeriod,
	}
}

// DefaultDepositParams default parameters for deposits
func DefaultDepositParams() DepositParams {
	return NewDepositParams(
		sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, DefaultMinDepositTokens)),
		DefaultPeriod,
	)
}

// String implements stringer insterface
func (dp DepositParams) String() string {
	out, _ := yaml.Marshal(dp)
	return string(out)
}

// Equal checks equality of DepositParams
func (dp DepositParams) Equal(dp2 DepositParams) bool {
	return dp.MinDeposit.IsEqual(dp2.MinDeposit) && dp.MaxDepositPeriod == dp2.MaxDepositPeriod
}

func validateDepositParams(i interface{}) error {
	v, ok := i.(DepositParams)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if !v.MinDeposit.IsValid() {
		return fmt.Errorf("invalid minimum deposit: %s", v.MinDeposit)
	}
	if v.MaxDepositPeriod <= 0 {
		return fmt.Errorf("maximum deposit period must be positive: %d", v.MaxDepositPeriod)
	}

	return nil
}

// NewTallyParams creates a new TallyParams object
func NewTallyParams(quorum, threshold, vetoThreshold sdk.Dec) TallyParams {
	return TallyParams{
		Quorum:        quorum,
		Threshold:     threshold,
		VetoThreshold: vetoThreshold,
	}
}

// DefaultTallyParams default parameters for tallying
func DefaultTallyParams() TallyParams {
	return NewTallyParams(DefaultQuorum, DefaultThreshold, DefaultVetoThreshold)
}

// Equal checks equality of TallyParams
func (tp TallyParams) Equal(other TallyParams) bool {
	return tp.Quorum.Equal(other.Quorum) && tp.Threshold.Equal(other.Threshold) && tp.VetoThreshold.Equal(other.VetoThreshold)
}

// String implements stringer insterface
func (tp TallyParams) String() string {
	out, _ := yaml.Marshal(tp)
	return string(out)
}

func validateTallyParams(i interface{}) error {
	v, ok := i.(TallyParams)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.Quorum.IsNegative() {
		return fmt.Errorf("quorom cannot be negative: %s", v.Quorum)
	}
	if v.Quorum.GT(sdk.OneDec()) {
		return fmt.Errorf("quorom too large: %s", v)
	}
	if !v.Threshold.IsPositive() {
		return fmt.Errorf("vote threshold must be positive: %s", v.Threshold)
	}
	if v.Threshold.GT(sdk.OneDec()) {
		return fmt.Errorf("vote threshold too large: %s", v)
	}
	if !v.VetoThreshold.IsPositive() {
		return fmt.Errorf("veto threshold must be positive: %s", v.Threshold)
	}
	if v.VetoThreshold.GT(sdk.OneDec()) {
		return fmt.Errorf("veto threshold too large: %s", v)
	}

	return nil
}

// NewVotingParams creates a new VotingParams object
func NewVotingParams(votingPeriod time.Duration) VotingParams {
	return VotingParams{
		VotingPeriod: votingPeriod,
	}
}

// DefaultVotingParams default parameters for voting
func DefaultVotingParams() VotingParams {
	return NewVotingParams(DefaultPeriod)
}

// Equal checks equality of TallyParams
func (vp VotingParams) Equal(other VotingParams) bool {
	return vp.VotingPeriod == other.VotingPeriod
}

// String implements stringer interface
func (vp VotingParams) String() string {
	out, _ := yaml.Marshal(vp)
	return string(out)
}

func validateVotingParams(i interface{}) error {
	v, ok := i.(VotingParams)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.VotingPeriod <= 0 {
		return fmt.Errorf("voting period must be positive: %s", v.VotingPeriod)
	}

	return nil
}

// Params returns all of the governance params
type Params struct {
	VotingParams  VotingParams  `json:"voting_params" yaml:"voting_params"`
	TallyParams   TallyParams   `json:"tally_params" yaml:"tally_params"`
	DepositParams DepositParams `json:"deposit_params" yaml:"deposit_params"`
}

func (gp Params) String() string {
	return gp.VotingParams.String() + "\n" +
		gp.TallyParams.String() + "\n" + gp.DepositParams.String()
}

// NewParams creates a new gov Params instance
func NewParams(vp VotingParams, tp TallyParams, dp DepositParams) Params {
	return Params{
		VotingParams:  vp,
		DepositParams: dp,
		TallyParams:   tp,
	}
}

// DefaultParams default governance params
func DefaultParams() Params {
	return NewParams(DefaultVotingParams(), DefaultTallyParams(), DefaultDepositParams())
}
