package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/subspace"
)

// Parameter store key
var (
	ParamStoreKeyDepositParams = []byte("depositparams")
	ParamStoreKeyVotingParams  = []byte("votingparams")
	ParamStoreKeyTallyParams   = []byte("tallyparams")
)

// Key declaration for parameters
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable(
		ParamStoreKeyDepositParams, DepositParams{},
		ParamStoreKeyVotingParams, VotingParams{},
		ParamStoreKeyTallyParams, TallyParams{},
	)
}

// Param around deposits for governance
type DepositParams struct {
	MinDeposit       sdk.Coins     `json:"min_deposit,omitempty"`        //  Minimum deposit for a proposal to enter voting period.
	MaxDepositPeriod time.Duration `json:"max_deposit_period,omitempty"` //  Maximum period for Atom holders to deposit on a proposal. Initial value: 2 months
}

// NewDepositParams creates a new DepositParams object
func NewDepositParams(minDeposit sdk.Coins, maxDepositPeriod time.Duration) DepositParams {
	return DepositParams{
		MinDeposit:       minDeposit,
		MaxDepositPeriod: maxDepositPeriod,
	}
}

func (dp DepositParams) String() string {
	return fmt.Sprintf(`Deposit Params:
  Min Deposit:        %s
  Max Deposit Period: %s`, dp.MinDeposit, dp.MaxDepositPeriod)
}

// Checks equality of DepositParams
func (dp DepositParams) Equal(dp2 DepositParams) bool {
	return dp.MinDeposit.IsEqual(dp2.MinDeposit) && dp.MaxDepositPeriod == dp2.MaxDepositPeriod
}

// Param around Tallying votes in governance
type TallyParams struct {
	Quorum    sdk.Dec `json:"quorum,omitempty"`    //  Minimum percentage of total stake needed to vote for a result to be considered valid
	Threshold sdk.Dec `json:"threshold,omitempty"` //  Minimum proportion of Yes votes for proposal to pass. Initial value: 0.5
	Veto      sdk.Dec `json:"veto,omitempty"`      //  Minimum value of Veto votes to Total votes ratio for proposal to be vetoed. Initial value: 1/3
}

// NewTallyParams creates a new TallyParams object
func NewTallyParams(quorum, threshold, veto sdk.Dec) TallyParams {
	return TallyParams{
		Quorum:    quorum,
		Threshold: threshold,
		Veto:      veto,
	}
}

func (tp TallyParams) String() string {
	return fmt.Sprintf(`Tally Params:
  Quorum:             %s
  Threshold:          %s
  Veto:               %s`,
		tp.Quorum, tp.Threshold, tp.Veto)
}

// Param around Voting in governance
type VotingParams struct {
	VotingPeriod time.Duration `json:"voting_period,omitempty"` //  Length of the voting period.
}

// NewVotingParams creates a new VotingParams object
func NewVotingParams(votingPeriod time.Duration) VotingParams {
	return VotingParams{
		VotingPeriod: votingPeriod,
	}
}

func (vp VotingParams) String() string {
	return fmt.Sprintf(`Voting Params:
  Voting Period:      %s`, vp.VotingPeriod)
}

// Params returns all of the governance params
type Params struct {
	VotingParams  VotingParams  `json:"voting_params"`
	TallyParams   TallyParams   `json:"tally_params"`
	DepositParams DepositParams `json:"deposit_params"`
}

func (gp Params) String() string {
	return gp.VotingParams.String() + "\n" +
		gp.TallyParams.String() + "\n" + gp.DepositParams.String()
}

func NewParams(vp VotingParams, tp TallyParams, dp DepositParams) Params {
	return Params{
		VotingParams:  vp,
		DepositParams: dp,
		TallyParams:   tp,
	}
}
