package tx

import (
	apibase "cosmossdk.io/api/cosmos/base/v1beta1"
	apitx "cosmossdk.io/api/cosmos/tx/v1beta1"
	"cosmossdk.io/client/v2/offchain"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ExtendedTxBuilder interface {
	SetExtensionOptions(...*codectypes.Any)
}

// TxBuilder defines an interface which an application-defined concrete transaction
// type must implement. Namely, it must be able to set messages, generate
// signatures, and provide canonical bytes to sign over. The transaction must
// also know how to encode itself.
type TxBuilder interface {
	GetTx() apitx.Tx
	SetMsgs(...sdk.Msg) error
	SetMemo(string)
	SetFeeAmount([]apibase.Coin)
	SetFeePayer(string)
	SetGasLimit(uint64)
	SetTimeoutHeight(uint64)
	SetFeeGranter(string)
	SetUnordered(bool)
	SetSignatures(...offchain.OffchainSignature) error
	SetAuxSignerData(apitx.AuxSignerData) error
}

type TxBuilderProvider interface {
	NewTxBuilder() TxBuilder
	WrapTxBuilder(apitx.Tx) (TxBuilder, error)
}
