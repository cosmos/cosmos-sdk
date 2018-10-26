package gov

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Procedure around Deposits for governance
type DepositProcedure struct {
	// Minimum deposit for a proposal to enter voting period.
	MinDeposit       sdk.Coins     `json:"min_deposit"`
	// Maximum period for Atom holders to deposit on a proposal. Initial value: 2 months
	MaxDepositPeriod time.Duration `json:"max_deposit_period"`
}

// Procedure around Tallying votes in governance
type TallyingProcedure struct {
	// Minimum propotion of Yes votes for proposal to pass. Initial value: 0.5
	Threshold         sdk.Dec `json:"threshold"`
	// Minimum value of Veto votes to Total votes ratio for proposal to be vetoed. Default: 1/3
	Veto              sdk.Dec `json:"veto"`
	// Penalty if validator does not vote
	GovernancePenalty sdk.Dec `json:"governance_penalty"`
}

// Procedure around Voting in governance
type VotingProcedure struct {
	VotingPeriod time.Duration `json:"voting_period"` //  Length of the voting period.
}
