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
	"github.com/cosmos/cosmos-sdk/types"
)

func Verify(ctx client.Context, digest []byte) error {
	tx := &apitx.Tx{}
	err := protojson.Unmarshal(digest, tx)
	if err != nil {
		return err
	}
	err = verify(ctx, tx)
	if err != nil {
		return err
	}
	fmt.Printf("Verification OK")
	return nil
}

func verify(ctx client.Context, tx *apitx.Tx) error {
	sigTx := builder{
		cdc: ctx.Codec,
		tx:  tx,
	}
	txConfig := ctx.TxConfig // TODO?
	signModeHandler := txConfig.SignModeHandler()
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
			sigAddr = types.AccAddress(pubKey.Address())
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

		txData := sigTx.GetSigningTxData()
		err = authsigning.VerifySignature(context.Background(), pubKey, txSignerData, sig.Data, signModeHandler, txData)
		if err != nil {
			return err
		}
	}
	return nil
}
