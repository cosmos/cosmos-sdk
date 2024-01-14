package tx

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
)

type ExtendedTxBuilder interface {
	SetExtensionOptions(extOpts ...*codectypes.Any)
}

// TxBuilder defines an interface which an application-defined concrete transaction
// type must implement. Namely, it must be able to set messages, generate
// signatures, and provide canonical bytes to sign over. The transaction must
// also know how to encode itself.
type TxBuilder interface {
	GetTx() TxV2
	SetMsgs(msgs ...sdk.MsgV2) error
	SetSignatures(signatures ...signingtypes.SignatureV2) error
	SetMemo(memo string)
	SetFeeAmount(amount sdk.Coins)
	SetFeePayer(feePayer sdk.AccAddress)
	SetGasLimit(limit uint64)
	SetTimeoutHeight(height uint64)
	SetFeeGranter(feeGranter sdk.AccAddress)
	AddAuxSignerData(tx.AuxSignerData) error
}
