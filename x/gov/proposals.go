package gov

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"strings"
)

//-----------------------------------------------------------
// Proposal interface
type Proposal interface {
	GetProposalID() int64
	SetProposalID(int64)

	GetTitle() string
	SetTitle(string)

	GetDescription() string
	SetDescription(string)

	GetProposalType() ProposalKind
	SetProposalType(ProposalKind)

	GetStatus() ProposalStatus
	SetStatus(ProposalStatus)

	GetSubmitBlock() int64
	SetSubmitBlock(int64)

	GetTotalDeposit() sdk.Coins
	SetTotalDeposit(sdk.Coins)

	GetVotingStartBlock() int64
	SetVotingStartBlock(int64)

	Execute(ctx sdk.Context, k Keeper) error
}

// checks if two proposals are equal
func ProposalEqual(proposalA Proposal, proposalB Proposal) bool {
	if proposalA.GetProposalID() != proposalB.GetProposalID() ||
		proposalA.GetTitle() != proposalB.GetTitle() ||
		proposalA.GetDescription() != proposalB.GetDescription() ||
		proposalA.GetProposalType() != proposalB.GetProposalType() ||
		proposalA.GetStatus() != proposalB.GetStatus() ||
		proposalA.GetSubmitBlock() != proposalB.GetSubmitBlock() ||
		!(proposalA.GetTotalDeposit().IsEqual(proposalB.GetTotalDeposit())) ||
		proposalA.GetVotingStartBlock() != proposalB.GetVotingStartBlock() {
		return false
	}
	return true
}

//-----------------------------------------------------------
// Text Proposals
type TextProposal struct {
	ProposalID   int64        `json:"proposal_id"`   //  ID of the proposal
	Title        string       `json:"title"`         //  Title of the proposal
	Description  string       `json:"description"`   //  Description of the proposal
	ProposalType ProposalKind `json:"proposal_type"` //  Type of proposal. Initial set {PlainTextProposal, SoftwareUpgradeProposal}

	Status ProposalStatus `json:"proposal_status"` //  Status of the Proposal {Pending, Active, Passed, Rejected}

	SubmitBlock  int64     `json:"submit_block"`  //  Height of the block where TxGovSubmitProposal was included
	TotalDeposit sdk.Coins `json:"total_deposit"` //  Current deposit on this proposal. Initial value is set at InitialDeposit

	VotingStartBlock int64 `json:"voting_start_block"` //  Height of the block where MinDeposit was reached. -1 if MinDeposit is not reached
}

// Implements Proposal Interface
var _ Proposal = (*TextProposal)(nil)

// nolint
func (tp TextProposal) GetProposalID() int64                       { return tp.ProposalID }
func (tp *TextProposal) SetProposalID(proposalID int64)            { tp.ProposalID = proposalID }
func (tp TextProposal) GetTitle() string                           { return tp.Title }
func (tp *TextProposal) SetTitle(title string)                     { tp.Title = title }
func (tp TextProposal) GetDescription() string                     { return tp.Description }
func (tp *TextProposal) SetDescription(description string)         { tp.Description = description }
func (tp TextProposal) GetProposalType() ProposalKind              { return tp.ProposalType }
func (tp *TextProposal) SetProposalType(proposalType ProposalKind) { tp.ProposalType = proposalType }
func (tp TextProposal) GetStatus() ProposalStatus                  { return tp.Status }
func (tp *TextProposal) SetStatus(status ProposalStatus)           { tp.Status = status }
func (tp TextProposal) GetSubmitBlock() int64                      { return tp.SubmitBlock }
func (tp *TextProposal) SetSubmitBlock(submitBlock int64)          { tp.SubmitBlock = submitBlock }
func (tp TextProposal) GetTotalDeposit() sdk.Coins                 { return tp.TotalDeposit }
func (tp *TextProposal) SetTotalDeposit(totalDeposit sdk.Coins)    { tp.TotalDeposit = totalDeposit }
func (tp TextProposal) GetVotingStartBlock() int64                 { return tp.VotingStartBlock }
func (tp *TextProposal) SetVotingStartBlock(votingStartBlock int64) {
	tp.VotingStartBlock = votingStartBlock
}
func (tp *TextProposal) Execute(ctx sdk.Context, k Keeper) error { return nil }

////////////////////  iris/cosmos-sdk begin  ///////////////////////////
type Op string

const (
	Add    Op = "add"
	Update Op = "update"
)

type Param struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Op    Op     `json:"op"`
}

type Params []Param

// Implements Proposal Interface
var _ Proposal = (*ParameterProposal)(nil)

type ParameterProposal struct {
	TextProposal
	Params Params `json:"params"`
}

func (pp *ParameterProposal) Execute(ctx sdk.Context, k Keeper) (err error) {

	logger := ctx.Logger().With("module", "x/gov")
	logger.Info("Execute ParameterProposal begin", "info", fmt.Sprintf("current height:%d", ctx.BlockHeight()))

	for _, data := range pp.Params {
		//param only begin with "gov/" can be update
		if !strings.HasPrefix(data.Key, Prefix) {
			errMsg := fmt.Sprintf("Parameter %s is not begin with %s", data.Key, Prefix)
			logger.Error("Execute ParameterProposal ", "err", errMsg)
			continue
		}
		if data.Op == Add {
			k.ps.GovSetter().Set(ctx, data.Key, data.Value)
		} else if data.Op == Update {
			bz := k.ps.GovSetter().GetRaw(ctx, data.Key)
			if bz == nil || len(bz) == 0 {
				logger.Error("Execute ParameterProposal ", "err", "Parameter "+data.Key+" is not exist")
			} else {
				k.ps.GovSetter().SetString(ctx, data.Key, data.Value)
			}
		}
	}
	return
}

var _ Proposal = (*SoftwareUpgradeProposal)(nil)

type SoftwareUpgradeProposal struct {
	TextProposal
}

func (sp *SoftwareUpgradeProposal) Execute(ctx sdk.Context, k Keeper) error {
	logger := ctx.Logger().With("module", "x/gov")
	logger.Info("Execute SoftwareProposal begin", "info", fmt.Sprintf("current height:%d", ctx.BlockHeight()))


	bz := k.ps.GovSetter().GetRaw(ctx, "upgrade/proposalId")
	if bz == nil || len(bz) == 0 {
		logger.Error("Execute SoftwareProposal ", "err", "Parameter upgrade/proposalId is not exist")
	} else {
		err := k.ps.GovSetter().Set(ctx, "upgrade/proposalId", sp.ProposalID)
		if err != nil {
			return err
		}
	}

	bz = k.ps.GovSetter().GetRaw(ctx, "upgrade/proposalAcceptHeight")
	if bz == nil || len(bz) == 0 {
		logger.Error("Execute SoftwareProposal ", "err", "Parameter upgrade/proposalAcceptHeight is not exist")
	} else {
		err := k.ps.GovSetter().Set(ctx, "upgrade/proposalAcceptHeight", ctx.BlockHeight())
		if err != nil {
			return err
		}
	}
	return nil
}

////////////////////  iris/cosmos-sdk end  ///////////////////////////

//-----------------------------------------------------------
// ProposalQueue
type ProposalQueue []int64

//-----------------------------------------------------------
// ProposalKind

// Type that represents Proposal Type as a byte
type ProposalKind byte

//nolint
const (
	ProposalTypeText            ProposalKind = 0x01
	ProposalTypeParameterChange ProposalKind = 0x02
	ProposalTypeSoftwareUpgrade ProposalKind = 0x03
)

// String to proposalType byte.  Returns ff if invalid.
func ProposalTypeFromString(str string) (ProposalKind, error) {
	switch str {
	case "Text":
		return ProposalTypeText, nil
	case "ParameterChange":
		return ProposalTypeParameterChange, nil
	case "SoftwareUpgrade":
		return ProposalTypeSoftwareUpgrade, nil
	default:
		return ProposalKind(0xff), errors.Errorf("'%s' is not a valid proposal type", str)
	}
}

// is defined ProposalType?
func validProposalType(pt ProposalKind) bool {
	if pt == ProposalTypeText ||
		pt == ProposalTypeParameterChange ||
		pt == ProposalTypeSoftwareUpgrade {
		return true
	}
	return false
}

// Marshal needed for protobuf compatibility
func (pt ProposalKind) Marshal() ([]byte, error) {
	return []byte{byte(pt)}, nil
}

// Unmarshal needed for protobuf compatibility
func (pt *ProposalKind) Unmarshal(data []byte) error {
	*pt = ProposalKind(data[0])
	return nil
}

// Marshals to JSON using string
func (pt ProposalKind) MarshalJSON() ([]byte, error) {
	return json.Marshal(pt.String())
}

// Unmarshals from JSON assuming Bech32 encoding
func (pt *ProposalKind) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return nil
	}

	bz2, err := ProposalTypeFromString(s)
	if err != nil {
		return err
	}
	*pt = bz2
	return nil
}

// Turns VoteOption byte to String
func (pt ProposalKind) String() string {
	switch pt {
	case ProposalTypeText:
		return "Text"
	case ProposalTypeParameterChange:
		return "ParameterChange"
	case ProposalTypeSoftwareUpgrade:
		return "SoftwareUpgrade"
	default:
		return ""
	}
}

// For Printf / Sprintf, returns bech32 when using %s
func (pt ProposalKind) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		s.Write([]byte(fmt.Sprintf("%s", pt.String())))
	default:
		s.Write([]byte(fmt.Sprintf("%v", byte(pt))))
	}
}

//-----------------------------------------------------------
// ProposalStatus

// Type that represents Proposal Status as a byte
type ProposalStatus byte

//nolint
const (
	StatusDepositPeriod ProposalStatus = 0x01
	StatusVotingPeriod  ProposalStatus = 0x02
	StatusPassed        ProposalStatus = 0x03
	StatusRejected      ProposalStatus = 0x04
)

// ProposalStatusToString turns a string into a ProposalStatus
func ProposalStatusFromString(str string) (ProposalStatus, error) {
	switch str {
	case "DepositPeriod":
		return StatusDepositPeriod, nil
	case "VotingPeriod":
		return StatusVotingPeriod, nil
	case "Passed":
		return StatusPassed, nil
	case "Rejected":
		return StatusRejected, nil
	default:
		return ProposalStatus(0xff), errors.Errorf("'%s' is not a valid proposal status", str)
	}
}

// is defined ProposalType?
func validProposalStatus(status ProposalStatus) bool {
	if status == StatusDepositPeriod ||
		status == StatusVotingPeriod ||
		status == StatusPassed ||
		status == StatusRejected {
		return true
	}
	return false
}

// Marshal needed for protobuf compatibility
func (status ProposalStatus) Marshal() ([]byte, error) {
	return []byte{byte(status)}, nil
}

// Unmarshal needed for protobuf compatibility
func (status *ProposalStatus) Unmarshal(data []byte) error {
	*status = ProposalStatus(data[0])
	return nil
}

// Marshals to JSON using string
func (status ProposalStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(status.String())
}

// Unmarshals from JSON assuming Bech32 encoding
func (status *ProposalStatus) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return nil
	}

	bz2, err := ProposalStatusFromString(s)
	if err != nil {
		return err
	}
	*status = bz2
	return nil
}

// Turns VoteStatus byte to String
func (status ProposalStatus) String() string {
	switch status {
	case StatusDepositPeriod:
		return "DepositPeriod"
	case StatusVotingPeriod:
		return "VotingPeriod"
	case StatusPassed:
		return "Passed"
	case StatusRejected:
		return "Rejected"
	default:
		return ""
	}
}

// For Printf / Sprintf, returns bech32 when using %s
func (status ProposalStatus) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		s.Write([]byte(fmt.Sprintf("%s", status.String())))
	default:
		s.Write([]byte(fmt.Sprintf("%v", byte(status))))
	}
}
