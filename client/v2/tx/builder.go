package tx

import (
	"cosmossdk.io/client/v2/offchain"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	typestx "github.com/cosmos/cosmos-sdk/types/tx"
)

type ExtendedTxBuilder interface {
	SetExtensionOptions(...*codectypes.Any)
}

// TxBuilder defines an interface which an application-defined concrete transaction
// type must implement. Namely, it must be able to set messages, generate
// signatures, and provide canonical bytes to sign over. The transaction must
// also know how to encode itself.
type TxBuilder interface {
	GetTx() typestx.Tx
	GetSigningTxData() offchain.TxData

	SetMsgs(...sdk.Msg) error
	SetMemo(string)
	SetFeeAmount([]sdk.Coin)
	SetFeePayer(string)
	SetGasLimit(uint64)
	SetTimeoutHeight(uint64)
	SetFeeGranter(string)
	SetUnordered(bool)
	SetSignatures(...offchain.OffchainSignature) error
	SetAuxSignerData(typestx.AuxSignerData) error
}

type TxBuilderProvider interface {
	NewTxBuilder() TxBuilder
	WrapTxBuilder(typestx.Tx) (TxBuilder, error)
}
