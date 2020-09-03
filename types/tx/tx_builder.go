package tx

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
)

// TxBuilder defines an interface which an application-defined concrete transaction
// type must implement. Namely, it must be able to set messages, generate
// signatures, and provide canonical bytes to sign over. The transaction must
// also know how to encode itself.
type TxBuilder interface {
	GetTx() signing.Tx
	// GetProtoTx returns the tx as a proto.Message.
	GetProtoTx() *Tx

	SetMsgs(msgs ...sdk.Msg) error
	SetSignatures(signatures ...signingtypes.SignatureV2) error
	SetMemo(memo string)
	SetFeeAmount(amount sdk.Coins)
	SetGasLimit(limit uint64)
	SetTimeoutHeight(height uint64)
}
