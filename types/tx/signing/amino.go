package signing

import (
	"fmt"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/multisig"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/types/tx"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func AminoJSONTxDecoder(aminoJsonMarshaler codec.JSONMarshaler, txGen context.TxGenerator) sdk.TxDecoder {
	return func(txBytes []byte) (sdk.Tx, error) {
		var aminoTx auth.StdTx
		err := aminoJsonMarshaler.UnmarshalJSON(txBytes, &aminoTx)
		if err != nil {
			return nil, err
		}

		txBuilder := txGen.NewTxBuilder()

		// set msgs
		err = txBuilder.SetMsgs(aminoTx.Msgs...)
		if err != nil {
			return nil, err
		}

		txBuilder.SetMemo(aminoTx.Memo)

		// set fee
		fee := txGen.NewFee()
		fee.SetGas(aminoTx.Fee.Gas)
		fee.SetAmount(aminoTx.Fee.Amount)
		err = txBuilder.SetFee(fee)
		if err != nil {
			return nil, err
		}

		n := len(aminoTx.Signatures)
		clientSigs := make([]context.ClientSignature, n)
		pubKeys := aminoTx.GetPubKeys()
		sigs := aminoTx.GetSignatures()

		for i := 0; i < n; i++ {
			csig := txGen.NewSignature()

			pubKey := pubKeys[i]

			miSig, ok := csig.(ModeInfoSignature)
			if !ok {
				return nil, fmt.Errorf("can't set ModeInfo")
			}
			miSig.SetModeInfo(makeAminoModeInfo(pubKey))

			fmt.Printf("amino decoded pubkey: %T\n", pubKey)
			err = csig.SetPubKey(pubKey)
			if err != nil {
				return nil, err
			}

			csig.SetSignature(sigs[i])

			clientSigs[i] = csig
		}

		err = txBuilder.SetSignatures(clientSigs...)
		if err != nil {
			return nil, err
		}

		return txBuilder.GetTx(), nil
	}
}

func makeAminoModeInfo(pk crypto.PubKey) *types.ModeInfo {
	if multisigPk, ok := pk.(multisig.PubKeyMultisigThreshold); ok {
		multi := &types.ModeInfo_Multi{}
		for i, k := range multisigPk.PubKeys {
			multi.ModeInfos[i] = makeAminoModeInfo(k)
		}
		return &types.ModeInfo{
			Sum: &types.ModeInfo_Multi_{
				Multi: multi,
			},
		}
	} else {
		return &types.ModeInfo{
			Sum: &types.ModeInfo_Single_{
				Single: &types.ModeInfo_Single{
					Mode: types.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
				},
			},
		}
	}
}
