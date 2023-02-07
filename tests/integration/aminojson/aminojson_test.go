package aminojson

import (
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
	"gotest.tools/v3/assert"
	"pgregory.net/rapid"

	authapi "cosmossdk.io/api/cosmos/auth/v1beta1"
	authzapi "cosmossdk.io/api/cosmos/authz/v1beta1"
	"cosmossdk.io/api/cosmos/crypto/ed25519"
	distapi "cosmossdk.io/api/cosmos/distribution/v1beta1"
	"cosmossdk.io/x/tx/aminojson"
	"cosmossdk.io/x/tx/rapidproto"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authztypes "github.com/cosmos/cosmos-sdk/x/authz"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

type equivalentType struct {
	pulsar      proto.Message
	gogo        gogoproto.Message
	anyTypeURLs []string
}

func et(gogo gogoproto.Message, pulsar proto.Message, anyTypes ...proto.Message) equivalentType {
	var anyTypeURLs []string
	for _, a := range anyTypes {
		anyTypeURLs = append(anyTypeURLs, fmt.Sprintf("/%s", a.ProtoReflect().Descriptor().FullName()))
	}
	return equivalentType{
		pulsar:      pulsar,
		gogo:        gogo,
		anyTypeURLs: anyTypeURLs,
	}
}

var equivTypes = []equivalentType{
	// auth
	et(&authtypes.Params{}, &authapi.Params{}),
	et(&authtypes.BaseAccount{}, &authapi.BaseAccount{}, &ed25519.PubKey{}),
}

func TestAminoJSON_Equivalence(t *testing.T) {
	encCfg := testutil.MakeTestEncodingConfig(auth.AppModuleBasic{})
	//aminoCdc := goamino.NewCodec()
	//protoCdc := codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
	//aj := aminojson.NewAminoJSON()

	//for _, tt := range equivTypes {
	//	desc := tt.pulsar.ProtoReflect().Descriptor()
	//	opts := desc.Options()
	//	if !proto.HasExtension(opts, amino.E_Name) {
	//		fmt.Printf("WARN: missing name extension for %s\n", desc.FullName())
	//		continue
	//	}
	//	name := proto.GetExtension(opts, amino.E_Name).(string)
	//	aminoCdc.RegisterConcrete(tt.gogo, name, nil)
	//}

	//params := &authapi.Params{}

	for _, tt := range equivTypes {
		genOpts := rapidproto.GeneratorOptions{
			AnyTypeURLs: tt.anyTypeURLs,
			Resolver:    protoregistry.GlobalTypes,
		}
		gen := rapidproto.MessageGenerator(tt.pulsar, genOpts)
		fmt.Printf("testing %s\n", tt.pulsar.ProtoReflect().Descriptor().FullName())
		rapid.Check(t, func(t *rapid.T) {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("Panic: %+v\n", r)
					t.FailNow()
				}
			}()
			msg := gen.Draw(t, "msg")
			protoBz, err := proto.Marshal(msg)
			require.NoError(t, err)
			err = encCfg.Codec.Unmarshal(protoBz, tt.gogo)
			require.NoError(t, err)
		})
	}
}

func TestAminoJSON_LegacyParity(t *testing.T) {
	cdc := goamino.NewCodec()
	cdc.RegisterConcrete(authtypes.Params{}, "cosmos-sdk/x/auth/Params", nil)
	cdc.RegisterConcrete(disttypes.MsgWithdrawDelegatorReward{}, "cosmos-sdk/MsgWithdrawDelegationReward", nil)
	cdc.RegisterConcrete(&ed25519.PubKey{}, cryptotypes.PubKeyName, nil)
	cdc.RegisterConcrete(&authtypes.ModuleAccount{}, "cosmos-sdk/ModuleAccount", nil)
	cdc.RegisterConcrete(&authtypes.MsgUpdateParams{}, "cosmos-sdk/x/auth/MsgUpdateParams", nil)
	cdc.RegisterConcrete(&authztypes.MsgGrant{}, "cosmos-sdk/MsgGrant", nil)
	cdc.RegisterConcrete(&authztypes.MsgExec{}, "cosmos-sdk/MsgExec", nil)

	aj := aminojson.NewAminoJSON()
	addr1 := types.AccAddress([]byte("addr1"))
	now := time.Now()

	cases := map[string]struct {
		gogo   gogoproto.Message
		pulsar proto.Message
	}{
		"auth/params": {gogo: &authtypes.Params{TxSigLimit: 10}, pulsar: &authapi.Params{TxSigLimit: 10}},
		"auth/module_account": {
			gogo:   &authtypes.ModuleAccount{BaseAccount: authtypes.NewBaseAccountWithAddress(addr1)},
			pulsar: &authapi.ModuleAccount{BaseAccount: &authapi.BaseAccount{Address: addr1.String()}},
		},
		"authz/msg_grant": {
			gogo:   &authztypes.MsgGrant{Grant: authztypes.Grant{Expiration: &now}},
			pulsar: &authzapi.MsgGrant{Grant: &authzapi.Grant{Expiration: timestamppb.New(now)}},
		},
		"authz/msg_grant/empty": {
			gogo:   &authztypes.MsgGrant{},
			pulsar: &authzapi.MsgGrant{Grant: &authzapi.Grant{}},
		},
		"authz/msg_update_params": {
			gogo:   &authtypes.MsgUpdateParams{Params: authtypes.Params{TxSigLimit: 10}},
			pulsar: &authapi.MsgUpdateParams{Params: &authapi.Params{TxSigLimit: 10}},
		},
		"authz/msg_exec": {
			gogo:   &authztypes.MsgExec{Msgs: []*codectypes.Any{}},
			pulsar: &authzapi.MsgExec{Msgs: []*anypb.Any{}},
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
			gogoBytes, err := cdc.MarshalJSON(tc.gogo)
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
	cdc := goamino.NewCodec()
	params := &authapi.Params{}
	cdc.RegisterConcrete(&authtypes.ModuleAccount{}, "cosmos-sdk/ModuleAccount", nil)
	cdc.RegisterConcrete(&authtypes.BaseAccount{}, "cosmos-sdk/BaseAccount", nil)
	cdc.RegisterConcrete(&authtypes.Params{}, "cosmos-sdk/Params", nil)
	paramsName := params.ProtoReflect().Descriptor().FullName()

	gen := rapidproto.MessageGenerator(&authapi.BaseAccount{}, rapidproto.GeneratorOptions{
		AnyTypeURLs: []string{string(paramsName)},
		Resolver:    protoregistry.GlobalTypes,
	})
	rapid.Check(t, func(t *rapid.T) {
		msg := gen.Draw(t, "msg")
		bz, err := aminojson.NewAminoJSON().MarshalAmino(msg)
		assert.NilError(t, err)
		gogobz, err := cdc.MarshalJSON(msg)
		require.NoError(t, err)
		require.Equal(t, string(gogobz), string(bz), "gogo: %s vs pulsar: %s", string(gogobz), string(bz))
	})
}
