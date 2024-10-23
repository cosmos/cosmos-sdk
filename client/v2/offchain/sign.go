package offchain

import (
	"context"
	"fmt"

	apisigning "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	"cosmossdk.io/client/v2/internal/account"
	"cosmossdk.io/client/v2/internal/offchain"
	clitx "cosmossdk.io/client/v2/tx"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/version"
	tx "github.com/cosmos/cosmos-sdk/x/auth/tx/config"
)

const (
	// ExpectedChainID defines the chain id an off-chain message must have
	ExpectedChainID = ""
	// ExpectedAccountNumber defines the account number an off-chain message must have
	ExpectedAccountNumber = 0
	// ExpectedSequence defines the sequence number an off-chain message must have
	ExpectedSequence = 0

	signMode = apisigning.SignMode_SIGN_MODE_TEXTUAL
)

// Sign signs given bytes using the specified encoder and SignMode.
func Sign(ctx client.Context, rawBytes []byte, fromName, encoding, output string) (string, error) {
	digest, err := encodeDigest(encoding, rawBytes)
	if err != nil {
		return "", err
	}

	keybase, err := keyring.NewAutoCLIKeyring(ctx.Keyring, ctx.AddressCodec)
	if err != nil {
		return "", err
	}

	txConfig, err := clitx.NewTxConfig(clitx.ConfigOptions{
		AddressCodec:               ctx.AddressCodec,
		Cdc:                        ctx.Codec,
		ValidatorAddressCodec:      ctx.ValidatorAddressCodec,
		EnablesSignModes:           []apisigning.SignMode{signMode},
		TextualCoinMetadataQueryFn: tx.NewGRPCCoinMetadataQueryFn(ctx),
	})
	if err != nil {
		return "", err
	}

	accRetriever := account.NewAccountRetriever(ctx.AddressCodec, ctx, ctx.InterfaceRegistry)

	params := clitx.TxParameters{
		ChainID:  ExpectedChainID,
		SignMode: signMode,
		AccountConfig: clitx.AccountConfig{
			AccountNumber: ExpectedAccountNumber,
			Sequence:      ExpectedSequence,
			FromName:      fromName,
		},
	}

	txf, err := clitx.NewFactory(keybase, ctx.Codec, accRetriever, txConfig, ctx.AddressCodec, ctx, params)
	if err != nil {
		return "", err
	}

	pubKey, err := keybase.GetPubKey(fromName)
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
