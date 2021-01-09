package v038

import (
	"fmt"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	v036gov "github.com/cosmos/cosmos-sdk/x/gov/legacy/v036"
)

const (
	// ModuleName is the name of this module
	ModuleName = "upgrade"

	// RouterKey is used to route governance proposals
	RouterKey = ModuleName

	// StoreKey is the prefix under which we store this module's data
	StoreKey = ModuleName

	// QuerierKey is used to handle abci_query requests
	QuerierKey = ModuleName
)

// Plan specifies information about a planned upgrade and when it should occur
type Plan struct {
	// Sets the name for the upgrade. This name will be used by the upgraded version of the software to apply any
	// special "on-upgrade" commands during the first BeginBlock method after the upgrade is applied. It is also used
	// to detect whether a software version can handle a given upgrade. If no upgrade handler with this name has been
	// set in the software, it will be assumed that the software is out-of-date when the upgrade Time or Height
	// is reached and the software will exit.
	Name string `json:"name,omitempty"`

	// The time after which the upgrade must be performed.
	// Leave set to its zero value to use a pre-defined Height instead.
	Time time.Time `json:"time,omitempty"`

	// The height at which the upgrade must be performed.
	// Only used if Time is not set.
	Height int64 `json:"height,omitempty"`

	// Any application specific upgrade info to be included on-chain
	// such as a git commit that validators could automatically upgrade to
	Info string `json:"info,omitempty"`
}

func (p Plan) String() string {
	due := p.DueAt()
	dueUp := strings.ToUpper(due[0:1]) + due[1:]
	return fmt.Sprintf(`Upgrade Plan
  Name: %s
  %s
  Info: %s`, p.Name, dueUp, p.Info)
}

// ValidateBasic does basic validation of a Plan
func (p Plan) ValidateBasic() error {
	if len(p.Name) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "name cannot be empty")
	}
	if p.Height < 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "height cannot be negative")
	}
	isValidTime := p.Time.Unix() > 0
	if !isValidTime && p.Height == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "must set either time or height")
	}
	if isValidTime && p.Height != 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "cannot set both time and height")
	}

	return nil
}

// ShouldExecute returns true if the Plan is ready to execute given the current context
func (p Plan) ShouldExecute(ctx sdk.Context) bool {
	if p.Time.Unix() > 0 {
		return !ctx.BlockTime().Before(p.Time)
	}
	if p.Height > 0 {
		return p.Height <= ctx.BlockHeight()
	}
	return false
}

// DueAt is a string representation of when this plan is due to be executed
func (p Plan) DueAt() string {
	if p.Time.Unix() > 0 {
		return fmt.Sprintf("time: %s", p.Time.UTC().Format(time.RFC3339))
	}
	return fmt.Sprintf("height: %d", p.Height)
}

const (
	ProposalTypeSoftwareUpgrade       string = "SoftwareUpgrade"
	ProposalTypeCancelSoftwareUpgrade string = "CancelSoftwareUpgrade"
)

// Software Upgrade Proposals
type SoftwareUpgradeProposal struct {
	Title       string `json:"title" yaml:"title"`
	Description string `json:"description" yaml:"description"`
	Plan        Plan   `json:"plan" yaml:"plan"`
}

func NewSoftwareUpgradeProposal(title, description string, plan Plan) v036gov.Content {
	return SoftwareUpgradeProposal{title, description, plan}
}

// Implements Proposal Interface
var _ v036gov.Content = SoftwareUpgradeProposal{}

func (sup SoftwareUpgradeProposal) GetTitle() string       { return sup.Title }
func (sup SoftwareUpgradeProposal) GetDescription() string { return sup.Description }
func (sup SoftwareUpgradeProposal) ProposalRoute() string  { return RouterKey }
func (sup SoftwareUpgradeProposal) ProposalType() string   { return ProposalTypeSoftwareUpgrade }
func (sup SoftwareUpgradeProposal) ValidateBasic() error {
	if err := sup.Plan.ValidateBasic(); err != nil {
		return err
	}
	return v036gov.ValidateAbstract(sup)
}

func (sup SoftwareUpgradeProposal) String() string {
	return fmt.Sprintf(`Software Upgrade Proposal:
  Title:       %s
  Description: %s
`, sup.Title, sup.Description)
}

// Cancel Software Upgrade Proposals
type CancelSoftwareUpgradeProposal struct {
	Title       string `json:"title" yaml:"title"`
	Description string `json:"description" yaml:"description"`
}

func NewCancelSoftwareUpgradeProposal(title, description string) v036gov.Content {
	return CancelSoftwareUpgradeProposal{title, description}
}

// Implements Proposal Interface
var _ v036gov.Content = CancelSoftwareUpgradeProposal{}

func (sup CancelSoftwareUpgradeProposal) GetTitle() string       { return sup.Title }
func (sup CancelSoftwareUpgradeProposal) GetDescription() string { return sup.Description }
func (sup CancelSoftwareUpgradeProposal) ProposalRoute() string  { return RouterKey }
func (sup CancelSoftwareUpgradeProposal) ProposalType() string {
	return ProposalTypeCancelSoftwareUpgrade
}
func (sup CancelSoftwareUpgradeProposal) ValidateBasic() error {
	return v036gov.ValidateAbstract(sup)
}

func (sup CancelSoftwareUpgradeProposal) String() string {
	return fmt.Sprintf(`Cancel Software Upgrade Proposal:
  Title:       %s
  Description: %s
`, sup.Title, sup.Description)
}

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(SoftwareUpgradeProposal{}, "cosmos-sdk/SoftwareUpgradeProposal", nil)
	cdc.RegisterConcrete(CancelSoftwareUpgradeProposal{}, "cosmos-sdk/CancelSoftwareUpgradeProposal", nil)
}
