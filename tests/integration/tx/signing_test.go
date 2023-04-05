package tx

import (
	"testing"

	"github.com/cosmos/cosmos-proto/rapidproto"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"pgregory.net/rapid"

	"cosmossdk.io/x/evidence"
	feegrantmodule "cosmossdk.io/x/feegrant/module"
	"cosmossdk.io/x/upgrade"
	"github.com/cosmos/cosmos-sdk/tests/integration/rapidgen"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"
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
				builder := encCfg.TxConfig.NewTxBuilder()
				builder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(100))))
			})
		})
	}
}
