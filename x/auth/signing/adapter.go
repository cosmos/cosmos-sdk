package signing

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/types/known/anypb"

	basev1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
	multisigv1beta1 "cosmossdk.io/api/cosmos/crypto/multisig/v1beta1"
	signingv1beta1 "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	txv1beta1 "cosmossdk.io/api/cosmos/tx/v1beta1"
	txsigning "cosmossdk.io/x/tx/signing"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
)

type V2AdaptableTx interface {
	Tx
	GetBodyBytes() []byte
	GetAuthInfoBytes() []byte
	GetSignerInfos() []*tx.SignerInfo
	GetTxBody() *tx.TxBody
}

// GetSignBytesAdapter returns the sign bytes for a given transaction and sign mode.  It accepts the arguments expected
// for signing in x/auth/tx and converts them to the arguments expected by the txsigning.HandlerMap, then applies
// HandlerMap.GetSignBytes to get the sign bytes.
func GetSignBytesAdapter(
	ctx context.Context,
	encoder sdk.TxEncoder,
	handlerMap *txsigning.HandlerMap,
	mode signing.SignMode,
	signerData SignerData,
	tx sdk.Tx,
) ([]byte, error) {
	adaptableTx, ok := tx.(V2AdaptableTx)
	if !ok {
		return nil, fmt.Errorf("expected tx to be V2AdaptableTx, got %T", tx)
	}
	txData := AdaptableToTxData(adaptableTx)

	txSignMode, err := internalSignModeToAPI(mode)
	if err != nil {
		return nil, err
	}

	anyPk, err := codectypes.NewAnyWithValue(signerData.PubKey)
	if err != nil {
		return nil, err
	}

	txSignerData := txsigning.SignerData{
		ChainID:       signerData.ChainID,
		AccountNumber: signerData.AccountNumber,
		Sequence:      signerData.Sequence,
		Address:       signerData.Address,
		PubKey: &anypb.Any{
			TypeUrl: anyPk.TypeUrl,
			Value:   anyPk.Value,
		},
	}
	// Generate the bytes to be signed.
	return handlerMap.GetSignBytes(ctx, txSignMode, txSignerData, txData)
}

func AdaptableToTxData(adaptableTx V2AdaptableTx) txsigning.TxData {
	body := adaptableTx.GetTxBody()

	msgs := make([]*anypb.Any, len(body.Messages))
	for i, msg := range body.Messages {
		msgs[i] = &anypb.Any{
			TypeUrl: msg.TypeUrl,
			Value:   msg.Value,
		}
	}

	extOptions := make([]*anypb.Any, len(body.ExtensionOptions))
	for i, extOption := range body.ExtensionOptions {
		extOptions[i] = &anypb.Any{
			TypeUrl: extOption.TypeUrl,
			Value:   extOption.Value,
		}
	}

	nonCriticalExtOptions := make([]*anypb.Any, len(body.NonCriticalExtensionOptions))
	for i, extOption := range body.NonCriticalExtensionOptions {
		nonCriticalExtOptions[i] = &anypb.Any{
			TypeUrl: extOption.TypeUrl,
			Value:   extOption.Value,
		}
	}

	feeCoins := adaptableTx.GetFee()
	feeAmount := make([]*basev1beta1.Coin, len(feeCoins))
	for i, coin := range feeCoins {
		feeAmount[i] = &basev1beta1.Coin{
			Denom:  coin.Denom,
			Amount: coin.Amount.String(),
		}
	}

	var txTip *txv1beta1.Tip
	tip := adaptableTx.GetTip()
	if tip != nil {
		tipCoins := tip.GetAmount()
		tipAmount := make([]*basev1beta1.Coin, len(tipCoins))
		for i, coin := range tipCoins {
			tipAmount[i] = &basev1beta1.Coin{
				Denom:  coin.Denom,
				Amount: coin.Amount.String(),
			}
		}
		txTip = &txv1beta1.Tip{
			Amount: tipAmount,
			Tipper: tip.Tipper,
		}
	}

	signerInfos := adaptableTx.GetSignerInfos()
	txSignerInfos := make([]*txv1beta1.SignerInfo, len(signerInfos))
	for i, signerInfo := range signerInfos {
		modeInfo := &txv1beta1.ModeInfo{}
		AdaptModeInfo(signerInfo.ModeInfo, modeInfo)
		txSignerInfo := &txv1beta1.SignerInfo{
			PublicKey: &anypb.Any{
				TypeUrl: signerInfo.PublicKey.TypeUrl,
				Value:   signerInfo.PublicKey.Value,
			},
			Sequence: signerInfo.Sequence,
			ModeInfo: modeInfo,
		}
		txSignerInfos[i] = txSignerInfo
	}

	txAuthInfo := &txv1beta1.AuthInfo{
		SignerInfos: txSignerInfos,
		Fee: &txv1beta1.Fee{
			Amount:   feeAmount,
			GasLimit: adaptableTx.GetGas(),
			Payer:    adaptableTx.FeePayer().String(),
			Granter:  adaptableTx.FeeGranter().String(),
		},
		Tip: txTip,
	}

	txBody := &txv1beta1.TxBody{
		Messages:                    msgs,
		Memo:                        body.Memo,
		TimeoutHeight:               body.TimeoutHeight,
		ExtensionOptions:            extOptions,
		NonCriticalExtensionOptions: nonCriticalExtOptions,
	}
	return txsigning.TxData{
		AuthInfo:      txAuthInfo,
		AuthInfoBytes: adaptableTx.GetAuthInfoBytes(),
		Body:          txBody,
		BodyBytes:     adaptableTx.GetBodyBytes(),
	}
}

func AdaptModeInfo(legacy *tx.ModeInfo, res *txv1beta1.ModeInfo) {
	switch mi := legacy.Sum.(type) {
	case *tx.ModeInfo_Single_:
		res.Sum = &txv1beta1.ModeInfo_Single_{
			Single: &txv1beta1.ModeInfo_Single{
				Mode: signingv1beta1.SignMode(legacy.GetSingle().Mode),
			},
		}
	case *tx.ModeInfo_Multi_:
		modeInfos := make([]*txv1beta1.ModeInfo, len(legacy.GetMulti().ModeInfos))
		for i, modeInfo := range legacy.GetMulti().ModeInfos {
			AdaptModeInfo(modeInfo, modeInfos[i])
		}
		res.Sum = &txv1beta1.ModeInfo_Multi_{
			Multi: &txv1beta1.ModeInfo_Multi{
				Bitarray: &multisigv1beta1.CompactBitArray{
					Elems:           mi.Multi.Bitarray.Elems,
					ExtraBitsStored: mi.Multi.Bitarray.ExtraBitsStored,
				},
				ModeInfos: modeInfos,
			},
		}
	}
}
