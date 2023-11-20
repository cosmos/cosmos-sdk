package signing

import (
	"context"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	basev1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
	protosecp256k1 "cosmossdk.io/api/cosmos/crypto/secp256k1"
	protosecp256r1 "cosmossdk.io/api/cosmos/crypto/secp256r1"
	signingv1beta1 "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	txv1beta1 "cosmossdk.io/api/cosmos/tx/v1beta1"
	txsigning "cosmossdk.io/x/tx/signing"
	"cosmossdk.io/x/tx/signing/direct"

	"github.com/cosmos/cosmos-proto/anyutil"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256r1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
)

var (
	msg = []byte("I am become, the great destroyer of the worlds, and I have come to destroy all people. With the exception of you, all soldiers here on both sides will be slain")

	sk1    = secp256k1.GenPrivKey()
	pk1    = sk1.PubKey().(*secp256k1.PubKey)
	pk1Any = must(anyutil.New(&protosecp256k1.PubKey{Key: must(pk1.Marshal())}))

	sr1    = must(secp256r1.GenPrivKey())
	pr1    = sr1.PubKey().(*secp256r1.PubKey)
	pr1Any = must(anyutil.New(&protosecp256r1.PubKey{Key: must(pr1.Marshal())}))

	sigK1     = must(sk1.Sign(msg))
	sigK1Data = &signing.SingleSignatureData{
		SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
		Signature: sigK1,
	}
	sigR1     = must(sr1.Sign(msg))
	sigR1Data = &signing.SingleSignatureData{
		SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
		Signature: sigR1,
	}

	chainID = "cosmos-test"
	seqNo   = uint64(7)

	sdK1 = &signing.SingleSignatureData{
		SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
		Signature: sigK1,
	}

	signerDataK1 = txsigning.SignerData{
		ChainID:       chainID,
		AccountNumber: 1,
		Sequence:      seqNo,
		PubKey:        pk1Any,
	}

	sdR1 = &signing.SingleSignatureData{
		SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
		Signature: sigR1,
	}
	signerDataR1 = txsigning.SignerData{
		ChainID:       chainID,
		AccountNumber: 1,
		Sequence:      seqNo,
		PubKey:        pr1Any,
	}

	handlerMap = txsigning.NewHandlerMap(direct.SignModeHandler{})

	memo   = "benchmark-test"
	txBody = &txv1beta1.TxBody{
		Messages: []*anypb.Any{must(anyutil.New(&bankv1beta1.MsgSend{}))},
		Memo:     memo,
	}

	feePayerAddr = "feepayer"

	signerInfoK1 = []*txv1beta1.SignerInfo{
		{
			PublicKey: pk1Any,
			ModeInfo: &txv1beta1.ModeInfo{
				Sum: &txv1beta1.ModeInfo_Single_{
					Single: &txv1beta1.ModeInfo_Single{
						Mode: signingv1beta1.SignMode_SIGN_MODE_DIRECT_AUX,
					},
				},
			},
			Sequence: seqNo,
		},
	}
	authInfoK1 = &txv1beta1.AuthInfo{
		Fee: &txv1beta1.Fee{
			Amount:   []*basev1beta1.Coin{{Denom: "uatom", Amount: "1000"}},
			GasLimit: 20000,
			Payer:    feePayerAddr,
		},
		SignerInfos: signerInfoK1,
	}

	signerInfoR1 = []*txv1beta1.SignerInfo{
		{
			PublicKey: pr1Any,
			ModeInfo: &txv1beta1.ModeInfo{
				Sum: &txv1beta1.ModeInfo_Single_{
					Single: &txv1beta1.ModeInfo_Single{
						Mode: signingv1beta1.SignMode_SIGN_MODE_DIRECT_AUX,
					},
				},
			},
			Sequence: seqNo,
		},
	}
	authInfoR1 = &txv1beta1.AuthInfo{
		Fee: &txv1beta1.Fee{
			Amount:   []*basev1beta1.Coin{{Denom: "uatom", Amount: "1000"}},
			GasLimit: 20000,
			Payer:    feePayerAddr,
		},
		SignerInfos: signerInfoR1,
	}

	bodyBytes = must(proto.Marshal(txBody))

	txDataK1 = txsigning.TxData{
		Body:          txBody,
		AuthInfo:      authInfoK1,
		AuthInfoBytes: must(proto.Marshal(authInfoK1)),
		BodyBytes:     bodyBytes,
	}

	txDataR1 = txsigning.TxData{
		Body:          txBody,
		AuthInfo:      authInfoR1,
		AuthInfoBytes: must(proto.Marshal(authInfoR1)),
		BodyBytes:     bodyBytes,
	}
)

var sink any = nil
var cases = []struct {
	name       string
	pk         cryptotypes.PubKey
	signerData txsigning.SignerData
	signature  signing.SignatureData
	txData     txsigning.TxData
	wantErr    string
}{
	{
		name:       "secp256k1 good signature",
		pk:         pk1,
		signerData: signerDataK1,
		signature:  sdK1,
		txData:     txDataK1,
		wantErr:    "",
	},
	{
		name:       "secp256r1 good signature",
		pk:         pr1,
		signerData: signerDataR1,
		signature:  sdR1,
		txData:     txDataR1,
		wantErr:    "",
	},
	{
		name:       "secp256k1 mismatched signature",
		pk:         pk1,
		signerData: signerDataK1,
		signature:  sdR1,
		txData:     txDataK1,
		wantErr:    "unable to verify single signer signature",
	},
	{
		name:       "secp256r1 mismatched signature",
		pk:         pr1,
		signerData: signerDataR1,
		signature:  sdK1,
		txData:     txDataR1,
		wantErr:    "unable to verify single signer signature",
	},
}

func must[T any](res T, err error) T {
	if err != nil {
		panic(err)
	}
	return res
}

func BenchmarkVerifySignature(b *testing.B) {
	b.ReportAllocs()
	ctx := context.Background()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, tt := range cases {
			err := VerifySignature(ctx, tt.pk, tt.signerData, tt.signature, handlerMap, tt.txData)
			if g, w := err == nil, tt.wantErr == ""; g != w {
				b.Errorf("%q: error mismatch:\n\tgot:  %v\n\twant: %q", tt.name, err, tt.wantErr)
			} else if err != nil && !strings.Contains(err.Error(), tt.wantErr) {
				b.Errorf("%q: non-nil mismatch:\n\tgot:  %v\n\twant: %q", tt.name, err, tt.wantErr)
			}

			sink = tt.pk
		}
	}

	if sink == nil {
		b.Fatal("Benchmark did not run!")
	}

	sink = nil
}
