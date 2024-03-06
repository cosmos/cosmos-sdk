package tx

import (
	txv1beta1 "cosmossdk.io/api/cosmos/tx/v1beta1"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/gogoproto/types"
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
	//SignTx(Signer, Tx) error, txv1beta1.Tx

	SetMsgs(msgs ...*types.Any) error
	SetMemo(memo string)
	SetFeeAmount(amount txv1beta1.Fee)
	SetFeePayer(feePayer string)
	SetGasLimit(limit uint64)
	SetTimeoutHeight(height uint64)
	SetFeeGranter(feeGranter string)
	SetAuxSignerData(data txv1beta1.AuxSignerData) error
}
