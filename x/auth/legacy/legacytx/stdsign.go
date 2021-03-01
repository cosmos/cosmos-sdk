package legacytx

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
)

// StdSignDoc is replay-prevention structure.
// It includes the result of msg.GetSignBytes(),
// as well as the ChainID (prevent cross chain replay)
// and the Sequence numbers for each signature (prevent
// inchain replay and enforce tx ordering per account).
type StdSignDoc struct {
	AccountNumber uint64            `json:"account_number" yaml:"account_number"`
	Sequence      uint64            `json:"sequence" yaml:"sequence"`
	TimeoutHeight uint64            `json:"timeout_height,omitempty" yaml:"timeout_height"`
	ChainID       string            `json:"chain_id" yaml:"chain_id"`
	Memo          string            `json:"memo" yaml:"memo"`
	Fee           json.RawMessage   `json:"fee" yaml:"fee"`
	Msgs          []json.RawMessage `json:"msgs" yaml:"msgs"`
}

// StdSignBytes returns the bytes to sign for a transaction.
func StdSignBytes(chainID string, accnum, sequence, timeout uint64, fee StdFee, msgs []sdk.Msg, memo string) []byte {
	msgsBytes := make([]json.RawMessage, 0, len(msgs))
	for _, msg := range msgs {
		msgsBytes = append(msgsBytes, json.RawMessage(msg.GetSignBytes()))
	}

	bz, err := legacy.Cdc.MarshalJSON(StdSignDoc{
		AccountNumber: accnum,
		ChainID:       chainID,
		Fee:           json.RawMessage(fee.Bytes()),
		Memo:          memo,
		Msgs:          msgsBytes,
		Sequence:      sequence,
		TimeoutHeight: timeout,
	})
	if err != nil {
		panic(err)
	}

	return sdk.MustSortJSON(bz)
}

// Deprecated: StdSignature represents a sig
type StdSignature struct {
	cryptotypes.PubKey `json:"pub_key" yaml:"pub_key"` // optional
	Signature          []byte                          `json:"signature" yaml:"signature"`
}

// StdSignatureToSignatureV2 converts a StdSignature to a SignatureV2
func StdSignatureToSignatureV2(cdc *codec.LegacyAmino, sig StdSignature) (signing.SignatureV2, error) {
	pk := sig.GetPubKey()
	data, err := pubKeySigToSigData(cdc, pk, sig.Signature)
	if err != nil {
		return signing.SignatureV2{}, err
	}

	return signing.SignatureV2{
		PubKey: pk,
		Data:   data,
	}, nil
}

func pubKeySigToSigData(cdc *codec.LegacyAmino, key cryptotypes.PubKey, sig []byte) (signing.SignatureData, error) {
	multiPK, ok := key.(multisig.PubKey)
	if !ok {
		return &signing.SingleSignatureData{
			SignMode:  signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
			Signature: sig,
		}, nil
	}
	var multiSig multisig.AminoMultisignature
	err := cdc.UnmarshalBinaryBare(sig, &multiSig)
	if err != nil {
		return nil, err
	}

	sigs := multiSig.Sigs
	sigDatas := make([]signing.SignatureData, len(sigs))
	pubKeys := multiPK.GetPubKeys()
	bitArray := multiSig.BitArray
	n := multiSig.BitArray.Count()
	signatures := multisig.NewMultisig(n)
	sigIdx := 0
	for i := 0; i < n; i++ {
		if bitArray.GetIndex(i) {
			data, err := pubKeySigToSigData(cdc, pubKeys[i], multiSig.Sigs[sigIdx])
			if err != nil {
				return nil, sdkerrors.Wrapf(err, "Unable to convert Signature to SigData %d", sigIdx)
			}

			sigDatas[sigIdx] = data
			multisig.AddSignature(signatures, data, sigIdx)
			sigIdx++
		}
	}

	return signatures, nil
}
