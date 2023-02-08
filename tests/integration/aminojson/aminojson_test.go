package aminojson

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
	goamino "github.com/tendermint/go-amino"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"pgregory.net/rapid"

	authapi "cosmossdk.io/api/cosmos/auth/v1beta1"
	authzapi "cosmossdk.io/api/cosmos/authz/v1beta1"
	"cosmossdk.io/api/cosmos/crypto/ed25519"
	distapi "cosmossdk.io/api/cosmos/distribution/v1beta1"
	"cosmossdk.io/x/tx/aminojson"
	"cosmossdk.io/x/tx/rapidproto"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authztypes "github.com/cosmos/cosmos-sdk/x/authz"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

type generatedType struct {
	pulsar proto.Message
	gogo   gogoproto.Message
	opts   rapidproto.GeneratorOptions
}

func genType(gogo gogoproto.Message, pulsar proto.Message, opts rapidproto.GeneratorOptions) generatedType {
	return generatedType{
		pulsar: pulsar,
		gogo:   gogo,
		opts:   opts,
	}
}

var (
	genOpts  = rapidproto.GeneratorOptions{Resolver: protoregistry.GlobalTypes}
	genTypes = []generatedType{
		// auth
		genType(&authtypes.Params{}, &authapi.Params{}, genOpts),
		genType(&authtypes.BaseAccount{}, &authapi.BaseAccount{}, genOpts.WithAnyTypes(&ed25519.PubKey{})),
		genType(&authtypes.ModuleAccount{}, &authapi.ModuleAccount{}, genOpts.WithAnyTypes(&ed25519.PubKey{})),
		genType(&authtypes.ModuleCredential{}, &authapi.ModuleCredential{}, genOpts),
		genType(&authtypes.MsgUpdateParams{}, &authapi.MsgUpdateParams{}, genOpts),

		// authz
		genType(&authztypes.GenericAuthorization{}, &authzapi.GenericAuthorization{}, genOpts),
		genType(&authztypes.Grant{}, &authzapi.Grant{},
			genOpts.WithAnyTypes(&authzapi.GenericAuthorization{}).
				WithDisallowNil().
				WithInterfaceHint("cosmos.authz.v1beta1.Authorization", &authzapi.GenericAuthorization{}),
		),
		genType(&authztypes.MsgGrant{}, &authzapi.MsgGrant{},
			genOpts.WithAnyTypes(&authzapi.GenericAuthorization{}).
				WithInterfaceHint("cosmos.authz.v1beta1.Authorization", &authzapi.GenericAuthorization{}).
				WithDisallowNil(),
		),
		genType(&authztypes.MsgExec{}, &authzapi.MsgExec{},
			genOpts.WithAnyTypes(&authzapi.MsgGrant{}, &authzapi.GenericAuthorization{}).
				WithDisallowNil().
				WithInterfaceHint("cosmos.authz.v1beta1.Authorization", &authzapi.GenericAuthorization{}).
				WithInterfaceHint("cosmos.base.v1beta1.Msg", &authzapi.MsgGrant{}),
		),
	}
)

func TestAminoJSON_Equivalence(t *testing.T) {
	encCfg := testutil.MakeTestEncodingConfig(
		auth.AppModuleBasic{}, authzmodule.AppModuleBasic{})
	aj := aminojson.NewAminoJSON()

	for _, tt := range genTypes {
		gen := rapidproto.MessageGenerator(tt.pulsar, tt.opts)
		fmt.Printf("testing %s\n", tt.pulsar.ProtoReflect().Descriptor().FullName())
		rapid.Check(t, func(t *rapid.T) {
			defer func() {
				if r := recover(); r != nil {
					//fmt.Printf("Panic: %+v\n", r)
					t.FailNow()
				}
			}()
			msg := gen.Draw(t, "msg")
			postFixPulsarMessage(msg)

			gogo := tt.gogo
			sanity := tt.pulsar

			protoBz, err := proto.Marshal(msg)
			require.NoError(t, err)

			err = proto.Unmarshal(protoBz, sanity)
			require.NoError(t, err)

			err = encCfg.Codec.Unmarshal(protoBz, gogo)
			require.NoError(t, err)

			legacyAminoJson, err := encCfg.Amino.MarshalJSON(gogo)
			aminoJson, err := aj.MarshalAmino(msg)
			if !bytes.Equal(legacyAminoJson, aminoJson) {
				println("UNMATCHED")
			}
			require.Equal(t, string(legacyAminoJson), string(aminoJson))
		})
	}
}

func TestAminoJSON_LegacyParity(t *testing.T) {
	encCfg := testutil.MakeTestEncodingConfig(auth.AppModuleBasic{}, authzmodule.AppModuleBasic{},
		distribution.AppModuleBasic{})

	aj := aminojson.NewAminoJSON()
	addr1 := types.AccAddress([]byte("addr1"))
	now := time.Now()

	genericAuth, _ := codectypes.NewAnyWithValue(&authztypes.GenericAuthorization{Msg: "foo"})
	genericAuthPulsar, _ := anypb.New(&authzapi.GenericAuthorization{Msg: "foo"})

	cases := map[string]struct {
		gogo   gogoproto.Message
		pulsar proto.Message
	}{
		"auth/params": {gogo: &authtypes.Params{TxSigLimit: 10}, pulsar: &authapi.Params{TxSigLimit: 10}},
		"auth/module_account": {
			gogo: &authtypes.ModuleAccount{BaseAccount: authtypes.NewBaseAccountWithAddress(addr1),
				Permissions: []string{}},
			pulsar: &authapi.ModuleAccount{BaseAccount: &authapi.BaseAccount{Address: addr1.String()}, Permissions: []string{}},
		},
		"authz/msg_grant": {
			gogo: &authztypes.MsgGrant{
				Grant: authztypes.Grant{Expiration: &now, Authorization: genericAuth}},
			pulsar: &authzapi.MsgGrant{
				Grant: &authzapi.Grant{Expiration: timestamppb.New(now), Authorization: genericAuthPulsar}},
		},
		"authz/msg_update_params": {
			gogo:   &authtypes.MsgUpdateParams{Params: authtypes.Params{TxSigLimit: 10}},
			pulsar: &authapi.MsgUpdateParams{Params: &authapi.Params{TxSigLimit: 10}},
		},
		"authz/msg_exec/empty_msgs": {
			gogo:   &authztypes.MsgExec{Msgs: []*codectypes.Any{}},
			pulsar: &authzapi.MsgExec{Msgs: []*anypb.Any{}},
		},
		"authz/msg_exec/null_msg": {
			gogo:   &authztypes.MsgExec{Msgs: []*codectypes.Any{(*codectypes.Any)(nil)}},
			pulsar: &authzapi.MsgExec{Msgs: []*anypb.Any{(*anypb.Any)(nil)}},
		},
		"distribution/delegator_starting_info": {
			gogo:   &disttypes.DelegatorStartingInfo{},
			pulsar: &distapi.DelegatorStartingInfo{},
		},
		"distribution/delegator_starting_info/non_zero_dec": {
			gogo:   &disttypes.DelegatorStartingInfo{Stake: types.NewDec(10)},
			pulsar: &distapi.DelegatorStartingInfo{Stake: "10.000000000000000000"},
		},
		"distribution/delegation_delegator_reward": {
			gogo:   &disttypes.DelegationDelegatorReward{},
			pulsar: &distapi.DelegationDelegatorReward{},
		},
		"distribution/community_pool_spend_proposal_with_deposit": {
			gogo:   &disttypes.CommunityPoolSpendProposalWithDeposit{},
			pulsar: &distapi.CommunityPoolSpendProposalWithDeposit{},
		},
		"distribution/msg_withdraw_delegator_reward": {
			gogo:   &disttypes.MsgWithdrawDelegatorReward{DelegatorAddress: "foo"},
			pulsar: &distapi.MsgWithdrawDelegatorReward{DelegatorAddress: "foo"},
		},
		"crypto/pubkey": {
			gogo: &cryptotypes.PubKey{Key: []byte("key")}, pulsar: &ed25519.PubKey{Key: []byte("key")},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			gogoBytes, err := encCfg.Amino.MarshalJSON(tc.gogo)
			require.NoError(t, err)

			pulsarBytes, err := aj.MarshalAmino(tc.pulsar)
			require.NoError(t, err)

			fmt.Printf("pulsar: %s\n", string(pulsarBytes))
			require.Equal(t, string(gogoBytes), string(pulsarBytes), "gogo: %s vs pulsar: %s", gogoBytes, pulsarBytes)
		})
	}
}

func TestScratch(t *testing.T) {
	ti := newTypeIndex(msgTypes)
	cdc := goamino.NewCodec()
	aj := aminojson.NewAminoJSON()

	msg := &authzapi.MsgExec{Msgs: []*anypb.Any{{TypeUrl: "", Value: nil}}}
	cdc.RegisterConcrete(&authztypes.MsgExec{}, "cosmos-sdk/MsgExec", nil)
	goMsg := &authztypes.MsgExec{}

	ti.deepClone(msg, goMsg)
	gobz, err := cdc.MarshalJSON(goMsg)
	require.NoError(t, err)
	bz, err := aj.MarshalAmino(msg)
	require.NoError(t, err)

	require.Equal(t, string(gobz), string(bz), "gogo: %s vs pulsar: %s", string(gobz), string(bz))

	fmt.Printf("gogo: %v\npulsar: %v\n", goMsg, msg)
}

func TestAny(t *testing.T) {
	cdc := goamino.NewCodec()
	a := &codectypes.Any{TypeUrl: "foo", Value: []byte("bar")}
	_, err := cdc.MarshalJSON(a)
	require.NoError(t, err)
}

func TestModuleAccount(t *testing.T) {
	encCfg := testutil.MakeTestEncodingConfig(auth.AppModuleBasic{}, authzmodule.AppModuleBasic{},
		distribution.AppModuleBasic{})

	aj := aminojson.NewAminoJSON()
	addr1 := types.AccAddress([]byte("addr1"))
	pulsar := &authapi.ModuleAccount{BaseAccount: &authapi.BaseAccount{Address: addr1.String()}, Permissions: []string{}}
	gogo := &authtypes.ModuleAccount{}

	protoBz, err := proto.Marshal(pulsar)
	require.NoError(t, err)
	err = encCfg.Codec.Unmarshal(protoBz, gogo)
	require.NoError(t, err)
	legacyAminoJson, err := encCfg.Amino.MarshalJSON(gogo)
	aminoJson, err := aj.MarshalAmino(pulsar)
	// yes this is expected.  empty []string is not the same as nil []string.  this is a bug in gogo.
	require.NotEqual(t, string(legacyAminoJson), string(aminoJson),
		"gogo: %s vs %s", string(legacyAminoJson), string(aminoJson))
}

func postFixPulsarMessage(msg proto.Message) {
	switch m := msg.(type) {
	case *authapi.ModuleAccount:
		if m.BaseAccount == nil {
			m.BaseAccount = &authapi.BaseAccount{}
		}
		_, _, bz := testdata.KeyTestPubAddr()
		text, _ := bech32.ConvertAndEncode("cosmos", bz)
		m.BaseAccount.Address = text

		// see negative test
		if len(m.Permissions) == 0 {
			m.Permissions = nil
		}
	case *authapi.MsgUpdateParams:
		// params is required in the gogo message
		if m.Params == nil {
			m.Params = &authapi.Params{}
		}
	case *authzapi.MsgGrant:
		// grant is required in the gogo message
		if m.Grant == nil {
			m.Grant = &authzapi.Grant{}
		}
	}
}
