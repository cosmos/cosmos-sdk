package signing

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec/legacy_global"
	multisig2 "github.com/cosmos/cosmos-sdk/crypto/multisig"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/types/tx"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func AminoJSONTxDecoder(aminoJsonMarshaler codec.JSONMarshaler, txGen client.TxGenerator) sdk.TxDecoder {
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
		txBuilder.SetFee(aminoTx.Fee.Amount)
		txBuilder.SetGasLimit(aminoTx.Fee.Gas)

		n := len(aminoTx.Signatures)
		clientSigs := make([]client.SignatureBuilder, n)
		sigs := aminoTx.Signatures

		for i := 0; i < n; i++ {
			data, err := stdSignatureToSignatureData(sigs[i])
			if err != nil {
				return nil, err
			}
			clientSigs[i] = client.SignatureBuilder{
				PubKey: sigs[i].GetPubKey(),
				Data:   data,
			}
		}

		err = txBuilder.SetSignatures(clientSigs...)
		if err != nil {
			return nil, err
		}

		return txBuilder.GetTx(), nil
	}
}

func stdSignatureToSignatureData(signature auth.StdSignature) (types.SignatureData, error) {
	pk := signature.GetPubKey()
	if multisigPk, ok := pk.(multisig2.MultisigPubKey); ok {
		pubKeys := multisigPk.GetPubKeys()
		var sig multisig2.AminoMultisignature
		err := legacy_global.Cdc.UnmarshalBinaryBare(signature.Signature, &sig)
		if err != nil {
			return nil, err
		}
		n := sig.BitArray.Size()
		datas := make([]types.SignatureData, n)
		sigIndex := 0
		for i := 0; i < n; i++ {
			if sig.BitArray.GetIndex(i) {
				datas[sigIndex], err = stdSignatureToSignatureData(auth.StdSignature{
					PubKey:    pubKeys[i].Bytes(),
					Signature: sig.Sigs[sigIndex],
				})
				if err != nil {
					return nil, err
				}
				sigIndex++
			}
		}
		return &types.MultiSignatureData{
			BitArray:   sig.BitArray,
			Signatures: datas[:sigIndex],
		}, nil
	} else {
		return &types.SingleSignatureData{
			SignMode:  types.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
			Signature: signature.Signature,
		}, nil
	}
}
