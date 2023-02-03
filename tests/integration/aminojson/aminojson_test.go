package aminojson

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
	goamino "github.com/tendermint/go-amino"
	"google.golang.org/protobuf/proto"
	"pgregory.net/rapid"

	"cosmossdk.io/api/amino"
	authapi "cosmossdk.io/api/cosmos/auth/v1beta1"
	authzapi "cosmossdk.io/api/cosmos/authz/v1beta1"
	"cosmossdk.io/api/cosmos/crypto/ed25519"
	distapi "cosmossdk.io/api/cosmos/distribution/v1beta1"
	"cosmossdk.io/x/tx/aminojson"
	"cosmossdk.io/x/tx/rapidproto"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authztypes "github.com/cosmos/cosmos-sdk/x/authz"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

type typeUnderTest struct {
	gogo   gogoproto.Message
	pulsar proto.Message
}

type typeIndex struct {
	gogoFields   map[string]map[string]reflect.StructField
	pulsarFields map[string]map[string]reflect.StructField
	pulsarToGogo map[string]string
}

var msgTypes = []typeUnderTest{
	// auth
	{gogo: &authtypes.Params{}, pulsar: &authapi.Params{}},
	{gogo: &authtypes.BaseAccount{}, pulsar: &authapi.BaseAccount{}},
	{gogo: &authtypes.ModuleAccount{}, pulsar: &authapi.ModuleAccount{}},
	// missing name extension, do we need it?
	// {gogo: &authtypes.ModuleCredential{}, pulsar: &authapi.ModuleCredential{}},
	{gogo: &authtypes.MsgUpdateParams{}, pulsar: &authapi.MsgUpdateParams{}},
	// authz
	{gogo: &authztypes.MsgGrant{}, pulsar: &authzapi.MsgGrant{}},
	{gogo: &authztypes.MsgRevoke{}, pulsar: &authzapi.MsgRevoke{}},
	{gogo: &authztypes.MsgExec{}, pulsar: &authzapi.MsgExec{}},
	{gogo: &authztypes.GenericAuthorization{}, pulsar: &authzapi.GenericAuthorization{}},
}

func fqTypeName(msg any) string {
	return fullyQualifiedTypeName(reflect.TypeOf(msg).Elem())
}

func fullyQualifiedTypeName(typ reflect.Type) string {
	pkgType := typ
	if typ.Kind() == reflect.Pointer || typ.Kind() == reflect.Slice || typ.Kind() == reflect.Map || typ.Kind() == reflect.Array {
		pkgType = typ.Elem()
	}
	pkgPath := pkgType.PkgPath()
	if pkgPath == "" {
		return fmt.Sprintf("%v", typ)
	}

	return fmt.Sprintf("%s/%v", pkgPath, typ)
}

func newTypeIndex(types []typeUnderTest) typeIndex {
	ti := typeIndex{
		gogoFields:   make(map[string]map[string]reflect.StructField),
		pulsarFields: make(map[string]map[string]reflect.StructField),
		pulsarToGogo: make(map[string]string),
	}
	for _, t := range types {
		gogoType := reflect.TypeOf(t.gogo).Elem()
		pulsarType := reflect.TypeOf(t.pulsar).Elem()

		ti.gogoFields[fqTypeName(t.gogo)] = make(map[string]reflect.StructField)
		ti.pulsarFields[fqTypeName(t.pulsar)] = make(map[string]reflect.StructField)
		ti.pulsarToGogo[fqTypeName(t.pulsar)] = fqTypeName(t.gogo)

		for i := 0; i < gogoType.NumField(); i++ {
			field := gogoType.Field(i)
			tag := field.Tag.Get("protobuf")
			if tag == "" {
				continue
			}
			n := strings.Split(tag, ",")[3]
			name := strings.Split(n, "=")[1]

			ti.gogoFields[fqTypeName(t.gogo)][name] = gogoType.Field(i)
		}
		for i := 0; i < pulsarType.NumField(); i++ {
			field := pulsarType.Field(i)
			tag := field.Tag.Get("protobuf")
			if tag == "" {
				continue
			}
			n := strings.Split(tag, ",")[3]
			name := strings.Split(n, "=")[1]
			ti.pulsarFields[fqTypeName(t.pulsar)][name] = pulsarType.Field(i)
		}
	}

	return ti
}

func (ti typeIndex) deepClone(pulsar proto.Message, gogo gogoproto.Message) {
	for n, pStructField := range ti.pulsarFields[fqTypeName(pulsar)] {
		gStructField := ti.gogoFields[fqTypeName(gogo)][n]
		pulsarField := reflect.ValueOf(pulsar).Elem().FieldByName(pStructField.Name)
		gogoField := reflect.ValueOf(gogo).Elem().FieldByName(gStructField.Name)
		ti.setField(pulsarField, gogoField)
	}
}

func (ti typeIndex) assertEquals(t *testing.T, pulsar proto.Message, gogo gogoproto.Message) {
	for n, pStructField := range ti.pulsarFields[fqTypeName(pulsar)] {
		gStructField := ti.gogoFields[fqTypeName(gogo)][n]
		pulsarField := reflect.ValueOf(pulsar).Elem().FieldByName(pStructField.Name)
		gogoField := reflect.ValueOf(gogo).Elem().FieldByName(gStructField.Name)
		ti.assertFieldEquals(t, pulsarField, gogoField)
	}
}

func (ti typeIndex) setField(pulsar reflect.Value, gogo reflect.Value) {
	switch pulsar.Type().Kind() {
	case reflect.Ptr:
		panic(fmt.Sprintf("pointer not supported: %s", pulsar.Type()))
	default:
		gogo.Set(pulsar)
	}
}

func (ti typeIndex) assertFieldEquals(t *testing.T, pulsarField reflect.Value, gogoField reflect.Value) {
	switch pulsarField.Type().Kind() {
	case reflect.Ptr:
		panic(fmt.Sprintf("pointer not supported: %s", pulsarField.Type()))
	default:
		require.Equal(t, pulsarField.Interface(), gogoField.Interface())
	}
}

func TestTypeIndex(t *testing.T) {
	ti := newTypeIndex(msgTypes)
	require.Equal(t, len(msgTypes), len(ti.gogoFields))
	require.Equal(t, len(msgTypes), len(ti.pulsarFields))
	for k, v := range ti.pulsarFields {
		require.Equal(t, len(v), len(ti.gogoFields[ti.pulsarToGogo[k]]), "failed on type: %s", k)
	}
}

func TestDeepClone(t *testing.T) {
	ti := newTypeIndex(msgTypes)
	tt := msgTypes[0]
	//	for _, tt := range msgTypes {
	gen := rapidproto.MessageGenerator(tt.pulsar, rapidproto.GeneratorOptions{})
	rapid.Check(t, func(rt *rapid.T) {
		msg := gen.Draw(rt, "msg").(proto.Message)
		ti.deepClone(msg, tt.gogo)
		ti.assertEquals(t, msg, tt.gogo)
	})
	//	}
}

func TestAminoJSON_AllTypes(t *testing.T) {
	cdc := goamino.NewCodec()
	for _, tt := range msgTypes {
		desc := tt.pulsar.ProtoReflect().Descriptor()
		opts := desc.Options()
		if !proto.HasExtension(opts, amino.E_Name) {
			panic(fmt.Sprintf("missing name extension for %s", desc.FullName()))
		}
		name := proto.GetExtension(opts, amino.E_Name).(string)
		cdc.RegisterConcrete(tt.gogo, name, nil)
	}

	// TODO
	// roundtrip the message into gogoproto, check equivalanece with pulsar

	//ti := newTypeIndex(msgTypes)
	for _, tt := range msgTypes {
		gen := rapidproto.MessageGenerator(tt.pulsar, rapidproto.GeneratorOptions{})
		fmt.Printf("testing %T\n", tt.pulsar)
		rapid.Check(t, func(t *rapid.T) {
			msg := gen.Draw(t, "msg")
			fmt.Printf("testing %T\n", msg)
			//
			//for k, gogoField := range gogoFieldsByTag {
			//	fmt.Printf("testing field %v\n", gogoField.Type())
			//	pulsarField := pulsarFieldsByTag[k]
			//	pulsarField.Set(gogoField)
			//}
		})
	}
}

func TestAminoJSON_LegacyParity(t *testing.T) {
	cdc := goamino.NewCodec()
	cdc.RegisterConcrete(authtypes.Params{}, "cosmos-sdk/x/auth/Params", nil)
	cdc.RegisterConcrete(disttypes.MsgWithdrawDelegatorReward{}, "cosmos-sdk/MsgWithdrawDelegationReward", nil)
	cdc.RegisterConcrete(&ed25519.PubKey{}, cryptotypes.PubKeyName, nil)
	aj := aminojson.NewAminoJSON()

	cases := map[string]struct {
		gogo   gogoproto.Message
		pulsar proto.Message
	}{
		"auth/params": {gogo: &authtypes.Params{TxSigLimit: 10}, pulsar: &authapi.Params{TxSigLimit: 10}},
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

			require.Equal(t, string(gogoBytes), string(pulsarBytes), "gogo: %s vs pulsar: %s", gogoBytes, pulsarBytes)
		})
	}
}
