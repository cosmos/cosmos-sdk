// DONTCOVER
package v036

import (
	"fmt"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v034gov "github.com/cosmos/cosmos-sdk/x/gov/legacy/v034"
)

const (
	ModuleName = "gov"
	RouterKey  = ModuleName

	ProposalTypeText string = "Text"

	MaxDescriptionLength int = 5000
	MaxTitleLength       int = 140
)

var (
	_ Content = TextProposal{}
)

type (
	Proposals     []Proposal
	ProposalQueue []uint64

	TextProposal struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	Content interface {
		GetTitle() string
		GetDescription() string
		ProposalRoute() string
		ProposalType() string
		ValidateBasic() error
		String() string
	}

	Proposal struct {
		Content `json:"content"`

		ProposalID       uint64                 `json:"id"`
		Status           v034gov.ProposalStatus `json:"proposal_status"`
		FinalTallyResult v034gov.TallyResult    `json:"final_tally_result"`

		SubmitTime     time.Time `json:"submit_time"`
		DepositEndTime time.Time `json:"deposit_end_time"`
		TotalDeposit   sdk.Coins `json:"total_deposit"`

		VotingStartTime time.Time `json:"voting_start_time"`
		VotingEndTime   time.Time `json:"voting_end_time"`
	}

	GenesisState struct {
		StartingProposalID uint64                `json:"starting_proposal_id"`
		Deposits           v034gov.Deposits      `json:"deposits"`
		Votes              v034gov.Votes         `json:"votes"`
		Proposals          []Proposal            `json:"proposals"`
		DepositParams      v034gov.DepositParams `json:"deposit_params"`
		VotingParams       v034gov.VotingParams  `json:"voting_params"`
		TallyParams        v034gov.TallyParams   `json:"tally_params"`
	}
)

func NewGenesisState(
	startingProposalID uint64, deposits v034gov.Deposits, votes v034gov.Votes, proposals []Proposal,
	depositParams v034gov.DepositParams, votingParams v034gov.VotingParams, tallyParams v034gov.TallyParams,
) GenesisState {

	return GenesisState{
		StartingProposalID: startingProposalID,
		Deposits:           deposits,
		Votes:              votes,
		Proposals:          proposals,
		DepositParams:      depositParams,
		VotingParams:       votingParams,
		TallyParams:        tallyParams,
	}
}

func NewTextProposal(title, description string) Content {
	return TextProposal{title, description}
}

func (tp TextProposal) GetTitle() string       { return tp.Title }
func (tp TextProposal) GetDescription() string { return tp.Description }
func (tp TextProposal) ProposalRoute() string  { return RouterKey }
func (tp TextProposal) ProposalType() string   { return ProposalTypeText }
func (tp TextProposal) ValidateBasic() error   { return ValidateAbstract(tp) }

func (tp TextProposal) String() string {
	return fmt.Sprintf(`Text Proposal:
  Title:       %s
  Description: %s
`, tp.Title, tp.Description)
}

func ErrInvalidProposalContent(msg string) error {
	return fmt.Errorf("invalid proposal content: %s", msg)
}

func ValidateAbstract(c Content) error {
	title := c.GetTitle()
	if len(strings.TrimSpace(title)) == 0 {
		return ErrInvalidProposalContent("proposal title cannot be blank")
	}
	if len(title) > MaxTitleLength {
		return ErrInvalidProposalContent(fmt.Sprintf("proposal title is longer than max length of %d", MaxTitleLength))
	}

	description := c.GetDescription()
	if len(description) == 0 {
		return ErrInvalidProposalContent("proposal description cannot be blank")
	}
	if len(description) > MaxDescriptionLength {
		return ErrInvalidProposalContent(fmt.Sprintf("proposal description is longer than max length of %d", MaxDescriptionLength))
	}

	return nil
}

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterInterface((*Content)(nil), nil)
	cdc.RegisterConcrete(TextProposal{}, "cosmos-sdk/TextProposal", nil)
}
