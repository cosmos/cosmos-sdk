package tx

import (
	txv1beta1 "cosmossdk.io/api/cosmos/tx/v1beta1"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"google.golang.org/protobuf/types/known/anypb"
)

type ExtendedTxBuilder interface {
	SetExtensionOptions(extOpts ...*codectypes.Any)
}

// TxBuilder defines an interface which an application-defined concrete transaction
// type must implement. Namely, it must be able to set messages, generate
// signatures, and provide canonical bytes to sign over. The transaction must
// also know how to encode itself.
type TxBuilder interface {
	GetTx() txv1beta1.Tx
	Sign() error
	SetMsgs(msgs ...*anypb.Any) error
	SetMemo(memo string)
	SetFeeAmount(amount txv1beta1.Fee)
	SetFeePayer(feePayer string)
	SetGasLimit(limit uint64)
	SetTimeoutHeight(height uint64)
	SetFeeGranter(feeGranter string)
	SetUnordered(v bool)
	SetSignatures(doc txv1beta1.SignDoc) error
	SetAuxSignerData(data txv1beta1.AuxSignerData) error
}

type TxBuilderProvider interface {
	NewTxBuilder() TxBuilder
	WrapTxBuilder(txv1beta1.Tx) (TxBuilder, error)
}
