package upgrade

import (
	"fmt"
	"time"
)
import sdk "github.com/cosmos/cosmos-sdk/types"

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

// Handler specifies the type of function that is called when an upgrade is applied
type Handler func(ctx sdk.Context, plan Plan)

func (plan Plan) String() string {
	var whenStr string
	if !plan.Time.IsZero() {
		whenStr = fmt.Sprintf("Time: %s", plan.Time.Format(time.RFC3339))
	} else {
		whenStr = fmt.Sprintf("Height: %d", plan.Height)
	}
	return fmt.Sprintf(`Upgrade Plan
  Name: %s
  %s
  Info: %s`, plan.Name, whenStr, plan.Info)
}
