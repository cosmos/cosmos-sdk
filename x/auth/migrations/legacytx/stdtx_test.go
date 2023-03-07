package legacytx

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	kmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
)

var (
	priv = ed25519.GenPrivKey()
	addr = sdk.AccAddress(priv.PubKey().Address())
)

func init() {
	amino := codec.NewLegacyAmino()
	RegisterLegacyAminoCodec(amino)
}

// Deprecated: use fee amount and gas limit separately on TxBuilder.
func NewTestStdFee() StdFee {
	return NewStdFee(100000,
		sdk.NewCoins(sdk.NewInt64Coin("atom", 150)),
	)
}

func TestStdSignBytes(t *testing.T) {
	type args struct {
		chainID       string
		accnum        uint64
		sequence      uint64
		timeoutHeight uint64
		fee           StdFee
		msgs          []sdk.Msg
		memo          string
		tip           *tx.Tip
	}
	defaultFee := NewTestStdFee()
	defaultTip := &tx.Tip{Tipper: addr.String(), Amount: sdk.NewCoins(sdk.NewInt64Coin("tiptoken", 150))}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"with timeout height",
			args{"1234", 3, 6, 10, defaultFee, []sdk.Msg{testdata.NewTestMsg(addr)}, "memo", nil},
			fmt.Sprintf(`{"account_number":"3","chain_id":"1234","fee":{"amount":[{"amount":"150","denom":"atom"}],"gas":"100000"},"memo":"memo","msgs":[["%s"]],"sequence":"6","timeout_height":"10"}`, addr),
		},
		{
			"no timeout height (omitempty)",
			args{"1234", 3, 6, 0, defaultFee, []sdk.Msg{testdata.NewTestMsg(addr)}, "memo", nil},
			fmt.Sprintf(`{"account_number":"3","chain_id":"1234","fee":{"amount":[{"amount":"150","denom":"atom"}],"gas":"100000"},"memo":"memo","msgs":[["%s"]],"sequence":"6"}`, addr),
		},
		{
			"empty fee",
			args{"1234", 3, 6, 0, StdFee{}, []sdk.Msg{testdata.NewTestMsg(addr)}, "memo", nil},
			fmt.Sprintf(`{"account_number":"3","chain_id":"1234","fee":{"amount":[],"gas":"0"},"memo":"memo","msgs":[["%s"]],"sequence":"6"}`, addr),
		},
		{
			"no fee payer and fee granter (both omitempty)",
			args{"1234", 3, 6, 0, StdFee{Amount: defaultFee.Amount, Gas: defaultFee.Gas}, []sdk.Msg{testdata.NewTestMsg(addr)}, "memo", nil},
			fmt.Sprintf(`{"account_number":"3","chain_id":"1234","fee":{"amount":[{"amount":"150","denom":"atom"}],"gas":"100000"},"memo":"memo","msgs":[["%s"]],"sequence":"6"}`, addr),
		},
		{
			"with fee granter, no fee payer (omitempty)",
			args{"1234", 3, 6, 0, StdFee{Amount: defaultFee.Amount, Gas: defaultFee.Gas, Granter: addr.String()}, []sdk.Msg{testdata.NewTestMsg(addr)}, "memo", nil},
			fmt.Sprintf(`{"account_number":"3","chain_id":"1234","fee":{"amount":[{"amount":"150","denom":"atom"}],"gas":"100000","granter":"%s"},"memo":"memo","msgs":[["%s"]],"sequence":"6"}`, addr, addr),
		},
		{
			"with fee payer, no fee granter (omitempty)",
			args{"1234", 3, 6, 0, StdFee{Amount: defaultFee.Amount, Gas: defaultFee.Gas, Payer: addr.String()}, []sdk.Msg{testdata.NewTestMsg(addr)}, "memo", nil},
			fmt.Sprintf(`{"account_number":"3","chain_id":"1234","fee":{"amount":[{"amount":"150","denom":"atom"}],"gas":"100000","payer":"%s"},"memo":"memo","msgs":[["%s"]],"sequence":"6"}`, addr, addr),
		},
		{
			"with fee payer and fee granter",
			args{"1234", 3, 6, 0, StdFee{Amount: defaultFee.Amount, Gas: defaultFee.Gas, Payer: addr.String(), Granter: addr.String()}, []sdk.Msg{testdata.NewTestMsg(addr)}, "memo", nil},
			fmt.Sprintf(`{"account_number":"3","chain_id":"1234","fee":{"amount":[{"amount":"150","denom":"atom"}],"gas":"100000","granter":"%s","payer":"%s"},"memo":"memo","msgs":[["%s"]],"sequence":"6"}`, addr, addr, addr),
		},
		{
			"no fee, with tip",
			args{"1234", 3, 6, 0, StdFee{}, []sdk.Msg{testdata.NewTestMsg(addr)}, "memo", defaultTip},
			fmt.Sprintf(`{"account_number":"3","chain_id":"1234","fee":{"amount":[],"gas":"0"},"memo":"memo","msgs":[["%s"]],"sequence":"6","tip":{"amount":[{"amount":"150","denom":"tiptoken"}],"tipper":"%s"}}`, addr, addr),
		},
		{
			"with fee and with tip",
			args{"1234", 3, 6, 0, StdFee{Amount: defaultFee.Amount, Gas: defaultFee.Gas, Payer: addr.String()}, []sdk.Msg{testdata.NewTestMsg(addr)}, "memo", defaultTip},
			fmt.Sprintf(`{"account_number":"3","chain_id":"1234","fee":{"amount":[{"amount":"150","denom":"atom"}],"gas":"100000","payer":"%s"},"memo":"memo","msgs":[["%s"]],"sequence":"6","tip":{"amount":[{"amount":"150","denom":"tiptoken"}],"tipper":"%s"}}`, addr, addr, addr),
		},
		{
			"with empty tip (but not nil), tipper cannot be empty",
			args{"1234", 3, 6, 0, defaultFee, []sdk.Msg{testdata.NewTestMsg(addr)}, "memo", &tx.Tip{Tipper: addr.String()}},
			fmt.Sprintf(`{"account_number":"3","chain_id":"1234","fee":{"amount":[{"amount":"150","denom":"atom"}],"gas":"100000"},"memo":"memo","msgs":[["%s"]],"sequence":"6","tip":{"amount":[],"tipper":"%s"}}`, addr, addr),
		},
	}
	for i, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := string(StdSignBytes(tc.args.chainID, tc.args.accnum, tc.args.sequence, tc.args.timeoutHeight, tc.args.fee, tc.args.msgs, tc.args.memo, tc.args.tip))
			require.Equal(t, tc.want, got, "Got unexpected result on test case i: %d", i)
		})
	}
}

func TestSignatureV2Conversions(t *testing.T) {
	_, pubKey, _ := testdata.KeyTestPubAddr()
	cdc := codec.NewLegacyAmino()
	sdk.RegisterLegacyAminoCodec(cdc)
	dummy := []byte("dummySig")
	sig := StdSignature{PubKey: pubKey, Signature: dummy}

	sigV2, err := StdSignatureToSignatureV2(cdc, sig)
	require.NoError(t, err)
	require.Equal(t, pubKey, sigV2.PubKey)
	require.Equal(t, &signing.SingleSignatureData{
		SignMode:  signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
		Signature: dummy,
	}, sigV2.Data)

	sigBz, err := SignatureDataToAminoSignature(cdc, sigV2.Data)
	require.NoError(t, err)
	require.Equal(t, dummy, sigBz)

	// multisigs
	_, pubKey2, _ := testdata.KeyTestPubAddr()
	multiPK := kmultisig.NewLegacyAminoPubKey(1, []cryptotypes.PubKey{
		pubKey, pubKey2,
	})
	dummy2 := []byte("dummySig2")
	bitArray := cryptotypes.NewCompactBitArray(2)
	bitArray.SetIndex(0, true)
	bitArray.SetIndex(1, true)
	msigData := &signing.MultiSignatureData{
		BitArray: bitArray,
		Signatures: []signing.SignatureData{
			&signing.SingleSignatureData{
				SignMode:  signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
				Signature: dummy,
			},
			&signing.SingleSignatureData{
				SignMode:  signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
				Signature: dummy2,
			},
		},
	}

	msig, err := SignatureDataToAminoSignature(cdc, msigData)
	require.NoError(t, err)

	sigV2, err = StdSignatureToSignatureV2(cdc, StdSignature{
		PubKey:    multiPK,
		Signature: msig,
	})
	require.NoError(t, err)
	require.Equal(t, multiPK, sigV2.PubKey)
	require.Equal(t, msigData, sigV2.Data)
}

func TestGetSignaturesV2(t *testing.T) {
	_, pubKey, _ := testdata.KeyTestPubAddr()
	dummy := []byte("dummySig")

	cdc := codec.NewLegacyAmino()
	sdk.RegisterLegacyAminoCodec(cdc)
	cryptocodec.RegisterCrypto(cdc)

	fee := NewStdFee(50000, sdk.Coins{sdk.NewInt64Coin("atom", 150)})
	sig := StdSignature{PubKey: pubKey, Signature: dummy}
	stdTx := NewStdTx([]sdk.Msg{testdata.NewTestMsg()}, fee, []StdSignature{sig}, "testsigs")

	sigs, err := stdTx.GetSignaturesV2()
	require.Nil(t, err)
	require.Equal(t, len(sigs), 1)

	require.Equal(t, cdc.MustMarshal(sigs[0].PubKey), cdc.MustMarshal(sig.GetPubKey()))
	require.Equal(t, sigs[0].Data, &signing.SingleSignatureData{
		SignMode:  signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
		Signature: sig.GetSignature(),
	})
}
