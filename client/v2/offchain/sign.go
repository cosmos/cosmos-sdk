package offchain

import (
	"context"
	"fmt"

	gogogrpc "github.com/cosmos/gogoproto/grpc"

	apisigning "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	clientcontext "cosmossdk.io/client/v2/context"
	"cosmossdk.io/client/v2/internal/account"
	"cosmossdk.io/client/v2/internal/offchain"
	clitx "cosmossdk.io/client/v2/tx"

	"github.com/cosmos/cosmos-sdk/version"
)

const (
	// ExpectedChainID defines the chain id an off-chain message must have
	ExpectedChainID = ""
	// ExpectedAccountNumber defines the account number an off-chain message must have
	ExpectedAccountNumber = 0
	// ExpectedSequence defines the sequence number an off-chain message must have
	ExpectedSequence = 0
)

var enabledSignModes = []apisigning.SignMode{
	apisigning.SignMode_SIGN_MODE_DIRECT,
	apisigning.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
}

// Sign signs given bytes using the specified encoder and SignMode.
func Sign(
	ctx clientcontext.Context,
	rawBytes []byte,
	conn gogogrpc.ClientConn,
	fromName, encoding, signMode, output string,
) (string, error) {
	digest, err := encodeDigest(encoding, rawBytes)
	if err != nil {
		return "", err
	}

	txConfig, err := clitx.NewTxConfig(clitx.ConfigOptions{
		AddressCodec:          ctx.AddressCodec,
		Cdc:                   ctx.Cdc,
		ValidatorAddressCodec: ctx.ValidatorAddressCodec,
		EnabledSignModes:      enabledSignModes,
	})
	if err != nil {
		return "", err
	}

	accRetriever := account.NewAccountRetriever(ctx.AddressCodec, conn, ctx.Cdc.InterfaceRegistry())

	sm, err := getSignMode(signMode)
	if err != nil {
		return "", err
	}
	params := clitx.TxParameters{
		ChainID:  ExpectedChainID,
		SignMode: sm,
		AccountConfig: clitx.AccountConfig{
			AccountNumber: ExpectedAccountNumber,
			Sequence:      ExpectedSequence,
			FromName:      fromName,
		},
	}

	txf, err := clitx.NewFactory(ctx.Keyring, ctx.Cdc, accRetriever, txConfig, ctx.AddressCodec, conn, params)
	if err != nil {
		return "", err
	}

	pubKey, err := ctx.Keyring.GetPubKey(fromName)
	if err != nil {
		return "", err
	}

	addr, err := ctx.AddressCodec.BytesToString(pubKey.Address())
	if err != nil {
		return "", err
	}

	msg := &offchain.MsgSignArbitraryData{
		AppDomain: version.AppName,
		Signer:    addr,
		Data:      digest,
	}

	signedTx, err := txf.BuildsSignedTx(context.Background(), msg)
	if err != nil {
		return "", err
	}

	bz, err := encode(output, signedTx, txConfig)
	if err != nil {
		return "", err
	}

	return string(bz), nil
}

func encode(output string, tx clitx.Tx, config clitx.TxConfig) ([]byte, error) {
	switch output {
	case "json":
		return config.TxJSONEncoder()(tx)
	case "text":
		return config.TxTextEncoder()(tx)
	default:
		return nil, fmt.Errorf("unsupported output type: %s", output)
	}
}

// getSignMode returns the corresponding apisigning.SignMode based on the provided mode string.
func getSignMode(mode string) (apisigning.SignMode, error) {
	switch mode {
	case "direct":
		return apisigning.SignMode_SIGN_MODE_DIRECT, nil
	case "amino-json":
		return apisigning.SignMode_SIGN_MODE_LEGACY_AMINO_JSON, nil
	}

	return apisigning.SignMode_SIGN_MODE_UNSPECIFIED, fmt.Errorf("unsupported sign mode: %s", mode)
}
