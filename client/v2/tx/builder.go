package tx

import (
	"time"

	base "cosmossdk.io/api/cosmos/base/v1beta1"
	apitx "cosmossdk.io/api/cosmos/tx/v1beta1"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/transaction"
	"google.golang.org/protobuf/types/known/anypb"
)

type txState struct {
	addressCodec address.Codec

	msgs             []transaction.Msg
	timeoutHeight    uint64
	timeoutTimestamp time.Time
	granter          []byte
	payer            []byte
	unordered        bool
	memo             string
	gasLimit         uint64
	fees             []*base.Coin
	signerInfos      []*apitx.SignerInfo
	signatures       [][]byte

	extensionOptions            []*anypb.Any
	nonCriticalExtensionOptions []*anypb.Any
}

// GetTx converts txBuilder messages to V2 and returns a Tx.
//func (b *txBuilder) GetTx() (Tx, error) {
//	return b.getTx()
//}

// getFee computes the transaction fee information for the txBuilder.
// It returns a pointer to an apitx.Fee struct containing the fee amount, gas limit, payer, and granter information.
// If the granter or payer addresses are set, it converts them from bytes to string using the addressCodec.
func (b *txState) getFee() (fee *apitx.Fee, err error) {
	granterStr := ""
	if b.granter != nil {
		granterStr, err = b.addressCodec.BytesToString(b.granter)
		if err != nil {
			return nil, err
		}
	}

	payerStr := ""
	if b.payer != nil {
		payerStr, err = b.addressCodec.BytesToString(b.payer)
		if err != nil {
			return nil, err
		}
	}

	fee = &apitx.Fee{
		Amount:   b.fees,
		GasLimit: b.gasLimit,
		Payer:    payerStr,
		Granter:  granterStr,
	}

	return fee, nil
}

// SetFeePayer sets the fee payer for the transaction.
func (b *txState) SetFeePayer(feePayer string) error {
	if feePayer == "" {
		return nil
	}

	addr, err := b.addressCodec.StringToBytes(feePayer)
	if err != nil {
		return err
	}
	b.payer = addr
	return nil
}

// SetFeeGranter sets the fee granter's address in the transaction builder.
// If the feeGranter string is empty, the function returns nil without setting an address.
// It converts the feeGranter string to bytes using the address codec and sets it as the granter address.
// Returns an error if the conversion fails.
func (b *txState) SetFeeGranter(feeGranter string) error {
	if feeGranter == "" {
		return nil
	}

	addr, err := b.addressCodec.StringToBytes(feeGranter)
	if err != nil {
		return err
	}
	b.granter = addr

	return nil
}
