package types

import (
	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type MockBaseAccount struct {
	Address       sdk.AccAddress `json:"address" yaml:"address"`
	Coins         sdk.Coins      `json:"coins" yaml:"coins"`
	PubKey        crypto.PubKey  `json:"public_key" yaml:"public_key"`
	AccountNumber uint64         `json:"account_number" yaml:"account_number"`
	Sequence      uint64         `json:"sequence" yaml:"sequence"`
}

var _ sdk.Tx = (*MockStdTx)(nil)

// MockStdTx implements sdk.Tx
type MockStdTx struct {
	Msgs []sdk.Msg  `json:"msg" yaml:"msg"`
	Fee  MockStdFee `json:"fee" yaml:"fee"`
	Memo string     `json:"memo" yaml:"memo"`
}

// GetMsgs returns all the transaction's messages
func (tx MockStdTx) GetMsgs() []sdk.Msg { return tx.Msgs }

// ValidateBasic implements sdk.Tx. Validation is not necessary as transactions do not need
// fees or signatures.
func (tx MockStdTx) ValidateBasic() sdk.Error {
	return nil
}

// MockStdFee
type MockStdFee struct {
	Amount sdk.Coins `json:"amount" yaml:"amount"`
	Gas    uint64    `json:"gas" yaml:"gas"`
}

// AuthDefaultTxDecoder logic for standard transaction decoding
func AuthDefaultTxDecoder(cdc *codec.Codec) sdk.TxDecoder {
	return func(txBytes []byte) (sdk.Tx, sdk.Error) {
		var tx = MockStdTx{}

		if len(txBytes) == 0 {
			return nil, sdk.ErrTxDecode("txBytes are empty")
		}

		// StdTx.Msg is an interface. The concrete types
		// are registered by MakeTxCodec
		err := cdc.UnmarshalBinaryLengthPrefixed(txBytes, &tx)
		if err != nil {
			return nil, sdk.ErrTxDecode("error decoding transaction").TraceSDK(err.Error())
		}

		return tx, nil
	}
}
