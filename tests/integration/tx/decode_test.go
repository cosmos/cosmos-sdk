package tx

import (
	"strings"
	"testing"

	"github.com/cosmos/cosmos-proto/rapidproto"
	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"pgregory.net/rapid"

	msgv1 "cosmossdk.io/api/cosmos/msg/v1"
	"cosmossdk.io/math"
	authzmodule "cosmossdk.io/x/authz/module"
	"cosmossdk.io/x/bank"
	"cosmossdk.io/x/consensus"
	"cosmossdk.io/x/distribution"
	"cosmossdk.io/x/evidence"
	feegrantmodule "cosmossdk.io/x/feegrant/module"
	"cosmossdk.io/x/gov"
	groupmodule "cosmossdk.io/x/group/module"
	"cosmossdk.io/x/mint"
	"cosmossdk.io/x/slashing"
	"cosmossdk.io/x/staking"
	"cosmossdk.io/x/tx/decode"
	txsigning "cosmossdk.io/x/tx/signing"
	"cosmossdk.io/x/upgrade"

	"github.com/cosmos/cosmos-sdk/codec/legacy"
	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/tests/integration/rapidgen"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
)

// TestDecode tests that the tx decoder can decode all the txs in the test suite.
func TestDecode(t *testing.T) {
	encCfg := testutil.MakeTestEncodingConfig(
		codectestutil.CodecOptions{}, auth.AppModule{}, authzmodule.AppModule{}, bank.AppModule{},
		consensus.AppModule{}, distribution.AppModule{}, evidence.AppModule{}, feegrantmodule.AppModule{},
		gov.AppModule{}, groupmodule.AppModule{}, mint.AppModule{},
		slashing.AppModule{}, staking.AppModule{}, upgrade.AppModule{}, vesting.AppModule{})
	legacytx.RegressionTestingAminoCodec = encCfg.Amino

	fee := sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(100)))
	gas := uint64(200)
	memo := "memo"
	accSeq := uint64(2)

	_, pubkey, _ := testdata.KeyTestPubAddr()
	anyPk, _ := codectypes.NewAnyWithValue(pubkey)
	var signerInfo []*txtypes.SignerInfo
	signerInfo = append(signerInfo, &txtypes.SignerInfo{
		PublicKey: anyPk,
		ModeInfo: &txtypes.ModeInfo{
			Sum: &txtypes.ModeInfo_Single_{
				Single: &txtypes.ModeInfo_Single{
					Mode: signing.SignMode_SIGN_MODE_DIRECT,
				},
			},
		},
		Sequence: accSeq,
	})

	authInfo := &txtypes.AuthInfo{
		Fee:         &txtypes.Fee{Amount: fee, GasLimit: gas},
		SignerInfos: signerInfo,
	}

	authInfoBytes, err := gogoproto.Marshal(authInfo)
	require.NoError(t, err)

	for _, tt := range rapidgen.SignableTypes {
		desc := tt.Pulsar.ProtoReflect().Descriptor()
		name := string(desc.FullName())
		t.Run(name, func(t *testing.T) {
			gen := rapidproto.MessageGenerator(tt.Pulsar, tt.Opts)
			rapid.Check(t, func(t *rapid.T) {
				msg := gen.Draw(t, "msg")
				signers, err := encCfg.TxConfig.SigningContext().GetSigners(msg)
				if len(signers) == 0 {
					return
				}
				if err != nil {
					if strings.Contains(err.Error(), "empty address string is not allowed") {
						return
					}
					require.NoError(t, err)
				}
				gogo := tt.Gogo
				sanity := tt.Pulsar

				protoBz, err := proto.Marshal(msg)
				require.NoError(t, err)

				err = proto.Unmarshal(protoBz, sanity)
				require.NoError(t, err)

				err = encCfg.Codec.Unmarshal(protoBz, gogo)
				require.NoError(t, err)

				txBuilder := encCfg.TxConfig.NewTxBuilder()

				sig := signing.SignatureV2{
					PubKey: pubkey,
					Data: &signing.SingleSignatureData{
						SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
						Signature: legacy.Cdc.MustMarshal(pubkey),
					},
					Sequence: accSeq,
				}

				require.True(t, proto.HasExtension(desc.Options(), msgv1.E_Signer))

				err = txBuilder.SetMsgs(tt.Gogo)
				require.NoError(t, err)
				txBuilder.SetFeeAmount(fee)
				txBuilder.SetGasLimit(gas)
				txBuilder.SetMemo(memo)
				err = txBuilder.SetSignatures(sig)
				require.NoError(t, err)

				tx := txBuilder.GetTx()
				txBytes, err := encCfg.TxConfig.TxEncoder()(tx)
				require.NoError(t, err)
				signContext, err := txsigning.NewContext(txsigning.Options{
					AddressCodec:          dummyAddressCodec{},
					ValidatorAddressCodec: dummyAddressCodec{},
				})
				require.NoError(t, err)
				decodeCtx, err := decode.NewDecoder(decode.Options{SigningContext: signContext})
				require.NoError(t, err)
				decodedTx, err := decodeCtx.Decode(txBytes)
				require.NoError(t, err)
				require.NotNil(t, decodedTx)

				require.Equal(t, authInfoBytes, decodedTx.TxRaw.AuthInfoBytes)

				anyGogoMsg, err := codectypes.NewAnyWithValue(tt.Gogo)
				require.NoError(t, err)

				txBody := &txtypes.TxBody{
					Memo: memo,
					Messages: []*codectypes.Any{
						anyGogoMsg,
					},
				}
				bodyBytes, err := gogoproto.Marshal(txBody)
				require.NoError(t, err)

				require.Equal(t, bodyBytes, decodedTx.TxRaw.BodyBytes)
			})
		})
	}

	legacytx.RegressionTestingAminoCodec = nil
}

type dummyAddressCodec struct{}

func (d dummyAddressCodec) StringToBytes(text string) ([]byte, error) {
	return []byte(text), nil
}

func (d dummyAddressCodec) BytesToString(bz []byte) (string, error) {
	return string(bz), nil
}
