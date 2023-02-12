package aminojson

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"pgregory.net/rapid"

	authapi "cosmossdk.io/api/cosmos/auth/v1beta1"
	authzapi "cosmossdk.io/api/cosmos/authz/v1beta1"
	bankapi "cosmossdk.io/api/cosmos/bank/v1beta1"
	v1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
	consensusapi "cosmossdk.io/api/cosmos/consensus/v1"
	"cosmossdk.io/api/cosmos/crypto/ed25519"
	distapi "cosmossdk.io/api/cosmos/distribution/v1beta1"
	govv1beta1 "cosmossdk.io/api/cosmos/gov/v1beta1"
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
	"github.com/cosmos/cosmos-sdk/x/bank"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/consensus"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
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

		// bank
		genType(&banktypes.MsgSend{}, &bankapi.MsgSend{}, genOpts.WithDisallowNil()),
		genType(&banktypes.MsgMultiSend{}, &bankapi.MsgMultiSend{}, genOpts.WithDisallowNil()),
		genType(&banktypes.MsgUpdateParams{}, &bankapi.MsgUpdateParams{}, genOpts.WithDisallowNil()),
		genType(&banktypes.MsgSetSendEnabled{}, &bankapi.MsgSetSendEnabled{}, genOpts),
		genType(&banktypes.SendAuthorization{}, &bankapi.SendAuthorization{}, genOpts),
		genType(&banktypes.Params{}, &bankapi.Params{}, genOpts),

		// consensus
		genType(&consensustypes.MsgUpdateParams{}, &consensusapi.MsgUpdateParams{}, genOpts.WithDisallowNil()),
	}
)

func TestAminoJSON_Equivalence(t *testing.T) {
	encCfg := testutil.MakeTestEncodingConfig(
		auth.AppModuleBasic{}, authzmodule.AppModuleBasic{}, bank.AppModuleBasic{}, consensus.AppModuleBasic{})
	aj := aminojson.NewAminoJSON()

	for _, tt := range genTypes {
		name := string(tt.pulsar.ProtoReflect().Descriptor().FullName())
		t.Run(name, func(t *testing.T) {
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
				require.NoError(t, err)
				aminoJson, err := aj.MarshalAmino(msg)
				require.NoError(t, err)
				//if !bytes.Equal(legacyAminoJson, aminoJson) {
				//	println("UNMATCHED")
				//}
				require.Equal(t, string(legacyAminoJson), string(aminoJson))
				//fmt.Printf("MATCHED %s\n", aminoJson)
			})
		})
	}
}

func newAny(t *testing.T, msg proto.Message) *anypb.Any {
	bz, err := proto.Marshal(msg)
	require.NoError(t, err)
	typeName := fmt.Sprintf("/%s", msg.ProtoReflect().Descriptor().FullName())
	return &anypb.Any{
		TypeUrl: typeName,
		Value:   bz,
	}
}

func TestAminoJSON_LegacyParity(t *testing.T) {
	encCfg := testutil.MakeTestEncodingConfig(auth.AppModuleBasic{}, authzmodule.AppModuleBasic{},
		bank.AppModuleBasic{}, distribution.AppModuleBasic{})

	aj := aminojson.NewAminoJSON()
	addr1 := types.AccAddress([]byte("addr1"))
	now := time.Now()

	genericAuth, _ := codectypes.NewAnyWithValue(&authztypes.GenericAuthorization{Msg: "foo"})
	genericAuthPulsar := newAny(t, &authzapi.GenericAuthorization{Msg: "foo"})

	cases := map[string]struct {
		gogo               gogoproto.Message
		pulsar             proto.Message
		pulsarMarshalFails bool

		// this will fail in cases where a lossy encoding of an empty array to protobuf occurs. the unmarshalled bytes
		// represent the array as nil, and a subsequent marshal to JSON represent the array as null instead of empty.
		roundTripUnequal bool

		// pulsar does not support marshalling a math.Dec as anything except a string.  Therefore, we cannot unmarshal
		// a pulsar encoded Math.dec (the string representation of a Decimal) into a gogo Math.dec (expecting an int64).
		protoUnmarshalFails bool
	}{
		"auth/params": {gogo: &authtypes.Params{TxSigLimit: 10}, pulsar: &authapi.Params{TxSigLimit: 10}},
		"auth/module_account": {
			gogo: &authtypes.ModuleAccount{
				BaseAccount: authtypes.NewBaseAccountWithAddress(addr1), Permissions: []string{}},
			pulsar: &authapi.ModuleAccount{
				BaseAccount: &authapi.BaseAccount{Address: addr1.String()}, Permissions: []string{}},
			roundTripUnequal: true,
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
		"distribution/delegator_starting_info": {
			gogo:   &disttypes.DelegatorStartingInfo{},
			pulsar: &distapi.DelegatorStartingInfo{},
		},
		"distribution/delegator_starting_info/non_zero_dec": {
			gogo:                &disttypes.DelegatorStartingInfo{Stake: types.NewDec(10)},
			pulsar:              &distapi.DelegatorStartingInfo{Stake: "10.000000000000000000"},
			protoUnmarshalFails: true,
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
		"consensus/evidence_params/duration": {
			gogo:   &govtypes.VotingParams{VotingPeriod: 1e9 + 7},
			pulsar: &govv1beta1.VotingParams{VotingPeriod: &durationpb.Duration{Seconds: 1, Nanos: 7}},
		},
		"consensus/evidence_params/big_duration": {
			gogo: &govtypes.VotingParams{VotingPeriod: time.Duration(rapidproto.MaxDurationSeconds*1e9) + 999999999},
			pulsar: &govv1beta1.VotingParams{VotingPeriod: &durationpb.Duration{
				Seconds: rapidproto.MaxDurationSeconds, Nanos: 999999999}},
		},
		"consensus/evidence_params/too_big_duration": {
			gogo: &govtypes.VotingParams{VotingPeriod: time.Duration(rapidproto.MaxDurationSeconds*1e9) + 999999999},
			pulsar: &govv1beta1.VotingParams{VotingPeriod: &durationpb.Duration{
				Seconds: rapidproto.MaxDurationSeconds + 1, Nanos: 999999999}},
			pulsarMarshalFails: true,
		},
		// amino.dont_omitempty + empty/nil lists produce some surprising results
		// TODO
		// custom renderer for coins
		"bank/send_authorization/empty_coins": {
			gogo:   &banktypes.SendAuthorization{SpendLimit: []types.Coin{}},
			pulsar: &bankapi.SendAuthorization{SpendLimit: []*v1beta1.Coin{}},
		},
		"bank/send_authorization/nil_coins": {
			gogo:   &banktypes.SendAuthorization{SpendLimit: nil},
			pulsar: &bankapi.SendAuthorization{SpendLimit: nil},
		},
		"bank/send_authorization/empty_list": {
			gogo:   &banktypes.SendAuthorization{AllowList: []string{}},
			pulsar: &bankapi.SendAuthorization{AllowList: []string{}},
		},
		"bank/send_authorization/nil_list": {
			gogo:   &banktypes.SendAuthorization{AllowList: nil},
			pulsar: &bankapi.SendAuthorization{AllowList: nil},
		},
		"bank/msg_multi_send/nil_everything": {
			gogo:   &banktypes.MsgMultiSend{},
			pulsar: &bankapi.MsgMultiSend{},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			gogoBytes, err := encCfg.Amino.MarshalJSON(tc.gogo)
			require.NoError(t, err)

			pulsarBytes, err := aj.MarshalAmino(tc.pulsar)
			if tc.pulsarMarshalFails {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			fmt.Printf("pulsar: %s\n", string(pulsarBytes))
			fmt.Printf("  gogo: %s\n", string(gogoBytes))
			require.Equal(t, string(gogoBytes), string(pulsarBytes))

			pulsarProtoBytes, err := proto.Marshal(tc.pulsar)
			require.NoError(t, err)

			gogoType := reflect.TypeOf(tc.gogo).Elem()
			newGogo := reflect.New(gogoType).Interface().(gogoproto.Message)

			err = encCfg.Codec.Unmarshal(pulsarProtoBytes, newGogo)
			if tc.protoUnmarshalFails {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			newGogoBytes, err := encCfg.Amino.MarshalJSON(newGogo)
			if tc.roundTripUnequal {
				require.NotEqual(t, string(gogoBytes), string(newGogoBytes))
				return
			}
			require.Equal(t, string(gogoBytes), string(newGogoBytes))
		})
	}
}

func TestScratch(t *testing.T) {
}

func TestSendAuthorization(t *testing.T) {
	encCfg := testutil.MakeTestEncodingConfig(auth.AppModuleBasic{}, authzmodule.AppModuleBasic{},
		distribution.AppModuleBasic{}, bank.AppModuleBasic{})

	aj := aminojson.NewAminoJSON()

	// beware, Coins has as custom MarshalJSON method which changes how nil is handled
	// nil -> [] (empty list)
	// []  -> [] (empty list)
	// https://github.com/cosmos/cosmos-sdk/blob/be9bd7a8c1b41b115d58f4e76ee358e18a52c0af/types/coin.go#L199

	// explicitly show the default for clarity
	pulsar := &bankapi.SendAuthorization{SpendLimit: []*v1beta1.Coin{}}
	sanityPulsar := &bankapi.SendAuthorization{}
	gogo := &banktypes.SendAuthorization{SpendLimit: types.Coins{}}

	protoBz, err := proto.Marshal(pulsar)
	require.NoError(t, err)

	err = encCfg.Codec.Unmarshal(protoBz, gogo)
	require.NoError(t, err)

	err = proto.Unmarshal(protoBz, sanityPulsar)

	// !!!
	//  empty []string is not the same as nil []string.  this is a bug in gogo.
	// `[]string` -> proto.Marshal -> legacyAmino.UnmarshalProto (unmarshals empty slice as nil)
	//    -> legacyAmino.MarshalJson -> `null`
	// `[]string` -> [proto.Marshal -> pulsar.Unmarshal] -> amino.MarshalJson -> `[]`
	require.Nil(t, gogo.SpendLimit)
	require.Nil(t, sanityPulsar.SpendLimit)
	require.NotNil(t, pulsar.SpendLimit)
	require.Zero(t, len(pulsar.SpendLimit))

	legacyAminoJson, err := encCfg.Amino.MarshalJSON(gogo)
	aminoJson, err := aj.MarshalAmino(sanityPulsar)

	require.Equal(t, string(legacyAminoJson), string(aminoJson))

	aminoJson, err = aj.MarshalAmino(pulsar)
	require.NoError(t, err)

	// at this point, pulsar.SpendLimit = [], and gogo.SpendLimit = nil, but they will both marshal to `[]`
	// this is *only* possible because of Cosmos SDK's custom MarshalJSON method for Coins
	require.Equal(t, string(legacyAminoJson), string(aminoJson))
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
