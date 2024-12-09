package offchain

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"google.golang.org/protobuf/types/known/anypb"

	clientcontext "cosmossdk.io/client/v2/context"
	clitx "cosmossdk.io/client/v2/tx"
	"cosmossdk.io/core/address"
	txsigning "cosmossdk.io/x/tx/signing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

// Verify verifies a digest after unmarshalling it.
func Verify(ctx clientcontext.Context, digest []byte, fileFormat string) error {
	txConfig, err := clitx.NewTxConfig(clitx.ConfigOptions{
		AddressCodec:          ctx.AddressCodec,
		Cdc:                   ctx.Cdc,
		ValidatorAddressCodec: ctx.ValidatorAddressCodec,
		EnabledSignModes:      enabledSignModes,
	})
	if err != nil {
		return err
	}

	dTx, err := unmarshal(fileFormat, digest, txConfig)
	if err != nil {
		return err
	}

	return verify(ctx.AddressCodec, txConfig, dTx)
}

// verify verifies given Tx.
func verify(addressCodec address.Codec, txConfig clitx.TxConfig, dTx clitx.Tx) error {
	signModeHandler := txConfig.SignModeHandler()

	signers, err := dTx.GetSigners()
	if err != nil {
		return err
	}

	sigs, err := dTx.GetSignatures()
	if err != nil {
		return err
	}

	if len(sigs) != len(signers) {
		return errors.New("mismatch between the number of signatures and signers")
	}

	for i, sig := range sigs {
		pubKey := sig.PubKey
		if !bytes.Equal(pubKey.Address(), signers[i]) {
			return errors.New("signature does not match its respective signer")
		}

		addr, err := addressCodec.BytesToString(pubKey.Address())
		if err != nil {
			return err
		}

		anyPk, err := codectypes.NewAnyWithValue(pubKey)
		if err != nil {
			return err
		}

		txSignerData := txsigning.SignerData{
			ChainID:       ExpectedChainID,
			AccountNumber: ExpectedAccountNumber,
			Sequence:      ExpectedSequence,
			Address:       addr,
			PubKey: &anypb.Any{
				TypeUrl: anyPk.TypeUrl,
				Value:   anyPk.Value,
			},
		}

		txData, err := dTx.GetSigningTxData()
		if err != nil {
			return err
		}

		err = verifySignature(context.Background(), pubKey, txSignerData, sig.Data, signModeHandler, txData)
		if err != nil {
			return err
		}
	}
	return nil
}

// unmarshal unmarshalls a digest to a Tx using protobuf protojson.
func unmarshal(format string, bz []byte, config clitx.TxConfig) (clitx.Tx, error) {
	switch format {
	case "json":
		return config.TxJSONDecoder()(bz)
	case "text":
		return config.TxTextDecoder()(bz)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// verifySignature verifies a transaction signature contained in SignatureData abstracting over different signing modes.
func verifySignature(
	ctx context.Context,
	pubKey cryptotypes.PubKey,
	signerData txsigning.SignerData,
	signatureData clitx.SignatureData,
	handler *txsigning.HandlerMap,
	txData txsigning.TxData,
) error {
	switch data := signatureData.(type) {
	case *clitx.SingleSignatureData:
		signBytes, err := handler.GetSignBytes(ctx, data.SignMode, signerData, txData)
		if err != nil {
			return err
		}
		if !pubKey.VerifySignature(signBytes, data.Signature) {
			return errors.New("unable to verify single signer signature")
		}
		return nil
	default:
		return fmt.Errorf("unexpected SignatureData %T", signatureData)
	}
}
