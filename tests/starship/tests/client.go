package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// CreditFromFaucet will request facuet of the chain for tokens to address
func CreditFromFaucet(config *Config, address string) error {
	url := fmt.Sprintf("%s/credit", config.GetChain(chainID).GetFaucetAddr())

	body := map[string]string{
		"address": address,
		"denom":   denom,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	_, err = http.Post(url, "application/json", bytes.NewBuffer(jsonBody)) //nolint // test url is dependent on the config file
	if err != nil {
		return err
	}

	// Check the response status code
	// Note: ignore response error due to cosmjs error
	// todo: add check for error
	// if res.StatusCode != http.StatusOK {
	//	return fmt.Errorf("request failed with status code: %d", res.StatusCode)
	//}

	return nil
}

// CreateTestTx creates a test tx with the given txConfig and txBuilder
func CreateTestTx(txConfig client.TxConfig, txBuilder client.TxBuilder, privs []cryptotypes.PrivKey, accNums, accSeqs []uint64, chainID string) (xauthsigning.Tx, []byte, error) {
	defaultSignMode, err := xauthsigning.APISignModeToInternal(txConfig.SignModeHandler().DefaultMode())
	if err != nil {
		return nil, nil, err
	}
	// First round: we gather all the signer infos. We use the "set empty
	// signature" hack to do that.
	var sigsV2 []signing.SignatureV2
	for i, priv := range privs {
		sigV2 := signing.SignatureV2{
			PubKey: priv.PubKey(),
			Data: &signing.SingleSignatureData{
				SignMode:  defaultSignMode,
				Signature: nil,
			},
			Sequence: accSeqs[i],
		}

		sigsV2 = append(sigsV2, sigV2)
	}
	err = txBuilder.SetSignatures(sigsV2...)
	if err != nil {
		return nil, nil, err
	}

	// Second round: all signer infos are set, so each signer can sign.
	sigsV2 = []signing.SignatureV2{}
	for i, priv := range privs {
		signerData := xauthsigning.SignerData{
			Address:       sdk.AccAddress(priv.PubKey().Bytes()).String(),
			ChainID:       chainID,
			AccountNumber: accNums[i],
			Sequence:      accSeqs[i],
			PubKey:        priv.PubKey(),
		}
		sigV2, err := tx.SignWithPrivKey(
			context.TODO(), defaultSignMode, signerData,
			txBuilder, priv, txConfig, accSeqs[i])
		if err != nil {
			return nil, nil, err
		}

		sigsV2 = append(sigsV2, sigV2)
	}
	err = txBuilder.SetSignatures(sigsV2...)
	if err != nil {
		return nil, nil, err
	}

	txBytes, err := txConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, nil, err
	}

	return txBuilder.GetTx(), txBytes, nil
}

// GetAccSeqNumber returns the account number and sequence number for the given address
func GetAccSeqNumber(grpcConn *grpc.ClientConn, address string) (uint64, uint64, error) {
	info, err := auth.NewQueryClient(grpcConn).AccountInfo(context.Background(), &auth.QueryAccountInfoRequest{Address: address})
	if err != nil {
		return 0, 0, err
	}
	return info.Info.GetAccountNumber(), info.Info.GetSequence(), nil
}
