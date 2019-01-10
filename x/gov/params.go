package gov

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GovParams is a common type for the different governance params
type GovParams interface {
	HumanReadableString() string
}

// Param around Deposits for governance
type DepositParams struct {
	MinDeposit       sdk.Coins     `json:"min_deposit"`        //  Minimum deposit for a proposal to enter voting period.
	MaxDepositPeriod time.Duration `json:"max_deposit_period"` //  Maximum period for Atom holders to deposit on a proposal. Initial value: 2 months
}

// HumanReadableString satisfies GovParams
func (dp DepositParams) HumanReadableString() string {
	return fmt.Sprintf(`Deposit Params:
  Min Deposit:         %s
  Max Deposit Period:  %s
`, dp.MinDeposit, dp.MaxDepositPeriod)
}

// Checks equality of DepositParams
func (dp DepositParams) Equal(dp2 DepositParams) bool {
	return dp.MinDeposit.IsEqual(dp2.MinDeposit) && dp.MaxDepositPeriod == dp2.MaxDepositPeriod
}

// Param around Tallying votes in governance
type TallyParams struct {
	Quorum            sdk.Dec `json:"quorum"`             //  Minimum percentage of total stake needed to vote for a result to be considered valid
	Threshold         sdk.Dec `json:"threshold"`          //  Minimum propotion of Yes votes for proposal to pass. Initial value: 0.5
	Veto              sdk.Dec `json:"veto"`               //  Minimum value of Veto votes to Total votes ratio for proposal to be vetoed. Initial value: 1/3
	GovernancePenalty sdk.Dec `json:"governance_penalty"` //  Penalty if validator does not vote
}

// HumanReadableString satisfies GovParams
func (tp TallyParams) HumanReadableString() string {
	return fmt.Sprintf(`Tally Params:
  Quorum:              %s
  Threshold:           %s
  Veto:                %s
  Goverance Penalty:   %s
`, tp.Quorum, tp.Threshold, tp.Veto, tp.GovernancePenalty)
}

// Param around Voting in governance
type VotingParams struct {
	VotingPeriod time.Duration `json:"voting_period"` //  Length of the voting period.
}

// HumanReadableString satisfies GovParams
func (vp VotingParams) HumanReadableString() string {
	return fmt.Sprintf(`Voting Params:
  Voting Period:       %s
`, vp.VotingPeriod)
}

// AllGovParams contains all the different params used by governance
type AllGovParams struct {
	DepositParams DepositParams `json:"deposit_params"`
	TallyParams   TallyParams   `json:"tally_params"`
	VotingParams  VotingParams  `json:"voting_params"`
}

// HumanReadableString satisfies GovParams
func (ap AllGovParams) HumanReadableString() string {
	return ap.DepositParams.HumanReadableString() +
		ap.TallyParams.HumanReadableString() +
		ap.VotingParams.HumanReadableString()
}
