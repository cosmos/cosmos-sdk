package offchain

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/anypb"

	apitx "cosmossdk.io/api/cosmos/tx/v1beta1"
	authsigning "cosmossdk.io/x/auth/signing"
	txsigning "cosmossdk.io/x/tx/signing"

	"github.com/cosmos/cosmos-sdk/client"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

func Verify(ctx client.Context, digest []byte) error {
	tx, err := unmarshall(digest)
	if err != nil {
		return err
	}

	return verify(ctx, tx)
}

func verify(ctx client.Context, tx *apitx.Tx) error {
	sigTx := builder{
		cdc: ctx.Codec,
		tx:  tx,
	}

	signModeHandler := ctx.TxConfig.SignModeHandler()

	signers, err := sigTx.GetSigners()
	if err != nil {
		return err
	}

	sigs, err := sigTx.GetSignatures()
	if err != nil {
		return err
	}

	if len(sigs) != len(signers) {
		return errors.New("")
	}

	for i, sig := range sigs {
		pubKey := sig.PubKey
		if !bytes.Equal(pubKey.Address(), signers[i]) {
			return errors.New("signature does not match its respective signer")
		}

		addr, err := ctx.AddressCodec.BytesToString(pubKey.Address())
		if err != nil {
			return err
		}

		signingData := authsigning.SignerData{
			Address:       addr,
			ChainID:       ExpectedChainID,
			AccountNumber: ExpectedAccountNumber,
			Sequence:      ExpectedSequence,
			PubKey:        pubKey,
		}

		anyPk, err := codectypes.NewAnyWithValue(pubKey)
		if err != nil {
			return err
		}

		txSignerData := txsigning.SignerData{
			ChainID:       signingData.ChainID,
			AccountNumber: signingData.AccountNumber,
			Sequence:      signingData.Sequence,
			Address:       signingData.Address,
			PubKey: &anypb.Any{
				TypeUrl: anyPk.TypeUrl,
				Value:   anyPk.Value,
			},
		}

		txData := sigTx.GetSigningTxData()
		err = verifySignature(context.Background(), pubKey, txSignerData, sig.Data, signModeHandler, txData)
		if err != nil {
			return err
		}
	}
	return nil
}

// unmarshall unmarshalls a digest to a Tx using protobuf protojson
func unmarshall(digest []byte) (*apitx.Tx, error) {
	tx := &apitx.Tx{}
	err := protojson.Unmarshal(digest, tx)
	return tx, err
}

// verifySignature verifies a transaction signature contained in SignatureData abstracting over different signing modes.
func verifySignature(
	ctx context.Context,
	pubKey cryptotypes.PubKey,
	signerData txsigning.SignerData,
	signatureData SignatureData,
	handler *txsigning.HandlerMap,
	txData txsigning.TxData,
) error {
	switch data := signatureData.(type) {
	case *SingleSignatureData:
		signBytes, err := handler.GetSignBytes(ctx, data.SignMode, signerData, txData)
		if err != nil {
			return err
		}
		if !pubKey.VerifySignature(signBytes, data.Signature) {
			return fmt.Errorf("unable to verify single signer signature")
		}
		return nil
	default:
		return fmt.Errorf("unexpected SignatureData %T", signatureData)
	}
}
