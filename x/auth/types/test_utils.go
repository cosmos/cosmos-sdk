package types

import (
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

func NewTestStdFee() StdFee {
	return NewStdFee(100000,
		sdk.NewCoins(sdk.NewInt64Coin("atom", 150)),
	)
}

// coins to more than cover the fee
func NewTestCoins() sdk.Coins {
	return sdk.Coins{
		sdk.NewInt64Coin("atom", 10000000),
	}
}

func KeyTestPubAddr() (crypto.PrivKey, crypto.PubKey, sdk.AccAddress) {
	key := secp256k1.GenPrivKey()
	pub := key.PubKey()
	addr := sdk.AccAddress(pub.Address())
	return key, pub, addr
}

func NewTestTx(ctx sdk.Context, msgs []sdk.Msg, privs []crypto.PrivKey, accNums []uint64, seqs []uint64, fee StdFee) sdk.Tx {
	sigs := make([]StdSignature, len(privs))
	for i, priv := range privs {
		signBytes := StdSignBytes(ctx.ChainID(), accNums[i], seqs[i], fee, msgs, "")

		sig, err := priv.Sign(signBytes)
		if err != nil {
			panic(err)
		}

		sigs[i] = StdSignature{PubKey: priv.PubKey().Bytes(), Signature: sig}
	}

	tx := NewStdTx(msgs, fee, sigs, "")
	return tx
}

// TODO Rename to NewTestTx
// TODO Is this impl better than using TxFactory?
func NewTestTx2(ctx sdk.Context, txGenerator client.TxGenerator, txBuilder client.TxBuilder, privs []crypto.PrivKey, accNums []uint64, seqs []uint64) ([]signing.SignatureV2, error) {
	var sigsV2 []signing.SignatureV2

	for i, priv := range privs {
		sigData := &signing.SingleSignatureData{
			SignMode:  txGenerator.SignModeHandler().DefaultMode(),
			Signature: nil,
		}
		sig := signing.SignatureV2{
			PubKey: priv.PubKey(),
			Data:   sigData,
		}

		err := txBuilder.SetSignatures(sig)
		if err != nil {
			return nil, err
		}

		signBytes, err := txGenerator.SignModeHandler().GetSignBytes(
			txGenerator.SignModeHandler().DefaultMode(),
			authsigning.SignerData{
				ChainID:         ctx.ChainID(),
				AccountNumber:   accNums[i],
				AccountSequence: seqs[i],
			},
			txBuilder.GetTx(),
		)
		if err != nil {
			return nil, err
		}

		sigBytes, err := priv.Sign(signBytes)
		if err != nil {
			return nil, err
		}

		sigData.Signature = sigBytes
		sig = signing.SignatureV2{
			PubKey: priv.PubKey(),
			Data:   sigData,
		}

		sigsV2 = append(sigsV2, sig)
	}

	return sigsV2, nil
}

func NewTestTxWithMemo(ctx sdk.Context, msgs []sdk.Msg, privs []crypto.PrivKey, accNums []uint64, seqs []uint64, fee StdFee, memo string) sdk.Tx {
	sigs := make([]StdSignature, len(privs))
	for i, priv := range privs {
		signBytes := StdSignBytes(ctx.ChainID(), accNums[i], seqs[i], fee, msgs, memo)

		sig, err := priv.Sign(signBytes)
		if err != nil {
			panic(err)
		}

		sigs[i] = StdSignature{PubKey: priv.PubKey().Bytes(), Signature: sig}
	}

	tx := NewStdTx(msgs, fee, sigs, memo)
	return tx
}

func NewTestTxWithSignBytes(msgs []sdk.Msg, privs []crypto.PrivKey, accNums []uint64, seqs []uint64, fee StdFee, signBytes []byte, memo string) sdk.Tx {
	sigs := make([]StdSignature, len(privs))
	for i, priv := range privs {
		sig, err := priv.Sign(signBytes)
		if err != nil {
			panic(err)
		}

		sigs[i] = StdSignature{PubKey: priv.PubKey().Bytes(), Signature: sig}
	}

	tx := NewStdTx(msgs, fee, sigs, memo)
	return tx
}
