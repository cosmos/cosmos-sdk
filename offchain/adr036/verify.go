package adr036

import (
	"bytes"
	"context"
	"errors"

	"google.golang.org/protobuf/types/known/anypb"

	txsigning "cosmossdk.io/x/tx/signing"

	"github.com/cosmos/cosmos-sdk/client"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

type OffChainVerifier struct {
	txConfig client.TxConfig
}

func (v OffChainVerifier) Verify(ctx context.Context, tx sdk.Tx) error {
	sigTx := tx.(authsigning.SigVerifiableTx)
	signModeHandler := v.txConfig.SignModeHandler()
	signers, err := sigTx.GetSigners()
	if err != nil {
		return err
	}

	sigs, err := sigTx.GetSignaturesV2()
	if err != nil {
		return err
	}

	if len(sigs) != len(signers) {
		return errors.New("")
	}

	for i, sig := range sigs {
		var (
			pubKey  = sig.PubKey
			sigAddr = sdk.AccAddress(pubKey.Address())
		)

		if !bytes.Equal(sigAddr, signers[i]) {
			return errors.New("signature does not match its respective signer")
		}

		signingData := authsigning.SignerData{
			Address:       sigAddr.String(),
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
		adaptableTx, ok := tx.(authsigning.V2AdaptableTx)
		if !ok {
			return errors.New("expexcted adaptable tx") // todo
		}
		txData := adaptableTx.GetSigningTxData()
		err = authsigning.VerifySignature(ctx, pubKey, txSignerData, sig.Data, signModeHandler, txData)
		if err != nil {
			return err
		}
	}
	return nil
}
