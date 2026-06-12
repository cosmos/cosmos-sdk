package tx

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	txsigning "github.com/cosmos/cosmos-sdk/x/tx/signing"
)

// GetSigningTxData returns an x/tx/signing.TxData representation of a transaction
// for use in the signing API defined in x/tx. Converts gogoproto wrapper types into
// the Go-native signing.TxData without any pulsar dependency.
func (w *wrapper) GetSigningTxData() txsigning.TxData {
	body := w.tx.Body
	authInfo := w.tx.AuthInfo

	msgs := make([]txsigning.RawMsg, len(body.Messages))
	for i, m := range body.Messages {
		msgs[i] = txsigning.RawMsg{TypeUrl: m.TypeUrl, Value: m.Value}
	}

	extOpts := make([]txsigning.RawMsg, len(body.ExtensionOptions))
	for i, e := range body.ExtensionOptions {
		extOpts[i] = txsigning.RawMsg{TypeUrl: e.TypeUrl, Value: e.Value}
	}

	nonCritExtOpts := make([]txsigning.RawMsg, len(body.NonCriticalExtensionOptions))
	for i, e := range body.NonCriticalExtensionOptions {
		nonCritExtOpts[i] = txsigning.RawMsg{TypeUrl: e.TypeUrl, Value: e.Value}
	}

	var ts *timestamppb.Timestamp
	if body.TimeoutTimestamp != nil {
		ts = timestamppb.New(*body.TimeoutTimestamp)
	}

	txBody := &txsigning.TxBodyData{
		Messages:                    msgs,
		Memo:                        body.Memo,
		TimeoutHeight:               body.TimeoutHeight,
		Unordered:                   body.Unordered,
		TimeoutTimestamp:            ts,
		ExtensionOptions:            extOpts,
		NonCriticalExtensionOptions: nonCritExtOpts,
	}

	feeCoins := make([]txsigning.TxCoinData, 0)
	if authInfo.Fee != nil {
		for _, c := range authInfo.Fee.Amount {
			feeCoins = append(feeCoins, txsigning.TxCoinData{
				Denom:  c.Denom,
				Amount: c.Amount.String(),
			})
		}
	}

	var feePayer, feeGranter string
	var gasLimit uint64
	if authInfo.Fee != nil {
		feePayer = authInfo.Fee.Payer
		feeGranter = authInfo.Fee.Granter
		gasLimit = authInfo.Fee.GasLimit
	}

	txAuthInfo := &txsigning.TxAuthInfoData{
		Fee: txsigning.TxFeeData{
			Amount:   feeCoins,
			GasLimit: gasLimit,
			Payer:    feePayer,
			Granter:  feeGranter,
		},
	}

	return txsigning.TxData{
		Body:          txBody,
		AuthInfo:      txAuthInfo,
		AuthInfoBytes: w.getAuthInfoBytes(),
		BodyBytes:     w.getBodyBytes(),
	}
}
