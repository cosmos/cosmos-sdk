package types

import (
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (p Plan) String() string {
	due := p.DueAt()
	dueUp := strings.ToUpper(due[0:1]) + due[1:]
	return fmt.Sprintf(`Upgrade Plan
  Name: %s
  %s
  Info: %s.`, p.Name, dueUp, p.Info)
}

// ValidateBasic does basic validation of a Plan
func (p Plan) ValidateBasic() error {
	if len(p.Name) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "name cannot be empty")
	}
	if p.Height < 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "height cannot be negative")
	}
	if p.Time.Unix() <= 0 && p.Height == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "must set either time or height")
	}
	if p.Time.Unix() > 0 && p.Height != 0 {
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
