package types

import (
	"fmt"
	"strings"
	"time"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	ibcexported "github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
)

var _ codectypes.UnpackInterfacesMessage = Plan{}

func (p Plan) String() string {
	due := p.DueAt()
	dueUp := strings.ToUpper(due[0:1]) + due[1:]
	var upgradedClientStr string
	upgradedClient, err := clienttypes.UnpackClientState(p.UpgradedClientState)
	if err != nil {
		upgradedClientStr = "no upgraded client provided"
	} else {
		upgradedClientStr = upgradedClient.String()
	}
	return fmt.Sprintf(`Upgrade Plan
  Name: %s
  %s
  Info: %s.
  Upgraded IBC Client: %s`, p.Name, dueUp, p.Info, upgradedClientStr)
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
	if p.Time.Unix() > 0 && p.UpgradedClientState != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "IBC chain upgrades must only set height")
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

// IsIBCPlan will return true if plan includes IBC client information
func (p Plan) IsIBCPlan() bool {
	return p.UpgradedClientState != nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (p Plan) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	// UpgradedClientState may be nil
	if p.UpgradedClientState == nil {
		return nil
	}

	var clientState ibcexported.ClientState
	return unpacker.UnpackAny(p.UpgradedClientState, &clientState)
}
