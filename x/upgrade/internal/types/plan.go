package types

import (
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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
	if p.Time.IsZero() && p.Height == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "must set either time or height")
	}
	if !p.Time.IsZero() && p.Height != 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "cannot set both time and height")
	}

	return nil
}

// ShouldExecute returns true if the Plan is ready to execute given the current context
func (p Plan) ShouldExecute(ctx sdk.Context) bool {
	if !p.Time.IsZero() {
		return !ctx.BlockTime().Before(p.Time)
	}
	if p.Height > 0 {
		return p.Height <= ctx.BlockHeight()
	}
	return false
}

// DueAt is a string representation of when this plan is due to be executed
func (p Plan) DueAt() string {
	if !p.Time.IsZero() {
		return fmt.Sprintf("time: %s", p.Time.UTC().Format(time.RFC3339))
	}
	return fmt.Sprintf("height: %d", p.Height)
}
