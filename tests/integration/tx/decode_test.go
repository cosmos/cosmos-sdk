package tx

import (
	"testing"

	"github.com/cosmos/cosmos-proto/rapidproto"
	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"pgregory.net/rapid"

	"cosmossdk.io/x/evidence"
	feegrantmodule "cosmossdk.io/x/feegrant/module"
	"cosmossdk.io/x/tx/decode"
	"cosmossdk.io/x/upgrade"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/tests/integration/rapidgen"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/consensus"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/gov"
	groupmodule "github.com/cosmos/cosmos-sdk/x/group/module"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

// TestDecode tests that the tx decoder can decode all the txs in the test suite.
func TestDecode(t *testing.T) {
	encCfg := testutil.MakeTestEncodingConfig(
		auth.AppModuleBasic{}, authzmodule.AppModuleBasic{}, bank.AppModuleBasic{}, consensus.AppModuleBasic{},
		distribution.AppModuleBasic{}, evidence.AppModuleBasic{}, feegrantmodule.AppModuleBasic{},
		gov.AppModuleBasic{}, groupmodule.AppModuleBasic{}, mint.AppModuleBasic{}, params.AppModuleBasic{},
		slashing.AppModuleBasic{}, staking.AppModuleBasic{}, upgrade.AppModuleBasic{}, vesting.AppModuleBasic{})

	fee := sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(100)))
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
		name := string(tt.Pulsar.ProtoReflect().Descriptor().FullName())
		t.Run(name, func(t *testing.T) {
			gen := rapidproto.MessageGenerator(tt.Pulsar, tt.Opts)
			rapid.Check(t, func(t *rapid.T) {
				msg := gen.Draw(t, "msg")
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

				gogoMsg, ok := gogo.(sdk.Msg)
				require.True(t, ok)

				err = txBuilder.SetMsgs(gogoMsg)
				require.NoError(t, err)
				txBuilder.SetFeeAmount(fee)
				txBuilder.SetGasLimit(gas)
				txBuilder.SetMemo(memo)
				err = txBuilder.SetSignatures(sig)
				require.NoError(t, err)

				tx := txBuilder.GetTx()
				txBytes, err := encCfg.TxConfig.TxEncoder()(tx)
				require.NoError(t, err)
				decodeCtx, err := decode.NewDecoder(decode.Options{})
				require.NoError(t, err)
				decodedTx, err := decodeCtx.Decode(txBytes)
				require.NoError(t, err)
				require.NotNil(t, decodedTx)

				require.Equal(t, authInfoBytes, decodedTx.TxRaw.AuthInfoBytes)

				anyGogoMsg, err := codectypes.NewAnyWithValue(gogoMsg)
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
}
