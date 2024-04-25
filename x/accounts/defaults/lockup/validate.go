package lockup

import (
	"cosmossdk.io/x/accounts/defaults/lockup/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func validateMsg(msg *types.MsgInitLockupAccount, isClawbackEnable bool) error {
	if msg.Owner == "" {
		return sdkerrors.ErrInvalidAddress.Wrap("account owner cannot be empty")
	}

	if isClawbackEnable && msg.Admin == "" {
		return sdkerrors.ErrInvalidAddress.Wrap("account admin cannot be empty")
	}

	return nil
}
