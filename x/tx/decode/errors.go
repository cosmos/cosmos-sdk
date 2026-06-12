package decode

import (
	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/x/tx/signing"
)

const txCodespace = "tx"

var (
	// ErrTxDecode is returned if we cannot parse a transaction.
	ErrTxDecode = errorsmod.Register(txCodespace, 1, "tx parse error")

	// ErrUnknownField is re-exported from x/tx/signing for backwards compatibility.
	ErrUnknownField = signing.ErrUnknownField
)
