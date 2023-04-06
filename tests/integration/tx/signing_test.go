package tx

import (
	"testing"

	"github.com/cosmos/cosmos-proto/anyutil"
	"github.com/cosmos/cosmos-proto/rapidproto"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"pgregory.net/rapid"

	"cosmossdk.io/x/evidence"
	feegrantmodule "cosmossdk.io/x/feegrant/module"
	"cosmossdk.io/x/tx/decode"
	"cosmossdk.io/x/upgrade"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	"github.com/cosmos/cosmos-sdk/tests/integration/rapidgen"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"
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

func TestDecode(t *testing.T) {
	encCfg := testutil.MakeTestEncodingConfig(
		auth.AppModuleBasic{}, authzmodule.AppModuleBasic{}, bank.AppModuleBasic{}, consensus.AppModuleBasic{},
		distribution.AppModuleBasic{}, evidence.AppModuleBasic{}, feegrantmodule.AppModuleBasic{},
		gov.AppModuleBasic{}, groupmodule.AppModuleBasic{}, mint.AppModuleBasic{}, params.AppModuleBasic{},
		slashing.AppModuleBasic{}, staking.AppModuleBasic{}, upgrade.AppModuleBasic{}, vesting.AppModuleBasic{})

	for _, tt := range rapidgen.DefaultGeneratedTypes {
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

				// TODO generate tx options
				txBuilder := encCfg.TxConfig.NewTxBuilder()
				fee := sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(100)))
				gas := uint64(200)
				memo := "memo"
				accSeq := uint64(2)

				_, pubkey, _ := testdata.KeyTestPubAddr()

				sig := signing.SignatureV2{
					PubKey: pubkey,
					Data: &signing.SingleSignatureData{
						SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
						Signature: legacy.Cdc.MustMarshal(pubkey),
					},
					Sequence: accSeq,
				}

				err = txBuilder.SetMsgs(gogo.(sdk.Msg))
				txBuilder.SetFeeAmount(fee)
				txBuilder.SetGasLimit(gas)
				txBuilder.SetMemo(memo)
				err = txBuilder.SetSignatures(sig)
				require.NoError(t, err)

				tx := txBuilder.GetTx()
				txBytes, err := encCfg.TxConfig.TxEncoder()(tx)
				anyutil.Unpack()
				decodeCtx, err := decode.NewContext(decode.Options{})
				require.NoError(t, err)
				decodedTx, err := decodeCtx.Decode(txBytes)
				require.NoError(t, err)
				require.Equal(t, tx.GetFee()[0].Amount, decodedTx.Tx.AuthInfo.Fee.Amount[0].Amount)
			})
		})
	}
}
