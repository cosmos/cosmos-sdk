package tx

import (
	gogoany "github.com/cosmos/gogoproto/types/any"

	base "cosmossdk.io/api/cosmos/base/v1beta1"
	typestx "cosmossdk.io/api/cosmos/tx/v1beta1"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/x/tx/signing"
)

type ExtendedTxBuilder interface {
	SetExtensionOptions(...*gogoany.Any) // TODO: sdk.Any?
}

// TxBuilder defines an interface which an application-defined concrete transaction
// type must implement. Namely, it must be able to set messages, generate
// signatures, and provide canonical bytes to sign over. The transaction must
// also know how to encode itself.
type TxBuilder interface {
	GetTx() typestx.Tx
	GetSigningTxData() signing.TxData

	SetMsgs(...transaction.Msg) error
	SetMemo(string)
	SetFeeAmount([]*base.Coin)
	SetFeePayer(string)
	SetGasLimit(uint64)
	SetTimeoutHeight(uint64)
	SetFeeGranter(string)
	SetUnordered(bool)
	SetSignatures(...Signature) error
	SetAuxSignerData(typestx.AuxSignerData) error
}

type TxBuilderProvider interface {
	NewTxBuilder() TxBuilder
	WrapTxBuilder(typestx.Tx) (TxBuilder, error)
}
