package aminojson

import (
	"bytes"
	"fmt"
	stdmath "math"
	"testing"
	"time"

	"github.com/cosmos/cosmos-proto/rapidproto"
	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"pgregory.net/rapid"

	authapi "cosmossdk.io/api/cosmos/auth/v1beta1"
	bankapi "cosmossdk.io/api/cosmos/bank/v1beta1"
	v1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
	msgv1 "cosmossdk.io/api/cosmos/msg/v1"
	"cosmossdk.io/math"
	authztypes "cosmossdk.io/x/authz"
	authzmodule "cosmossdk.io/x/authz/module"
	"cosmossdk.io/x/bank"
	banktypes "cosmossdk.io/x/bank/types"
	"cosmossdk.io/x/consensus"
	"cosmossdk.io/x/distribution"
	disttypes "cosmossdk.io/x/distribution/types"
	"cosmossdk.io/x/evidence"
	feegrantmodule "cosmossdk.io/x/feegrant/module"
	"cosmossdk.io/x/gov"
	gov_v1_types "cosmossdk.io/x/gov/types/v1"
	gov_v1beta1_types "cosmossdk.io/x/gov/types/v1beta1"
	groupmodule "cosmossdk.io/x/group/module"
	"cosmossdk.io/x/mint"
	"cosmossdk.io/x/slashing"
	slashingtypes "cosmossdk.io/x/slashing/types"
	"cosmossdk.io/x/staking"
	stakingtypes "cosmossdk.io/x/staking/types"
	"cosmossdk.io/x/tx/signing/aminojson"
	"cosmossdk.io/x/upgrade"

	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	ed25519types "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	secp256k1types "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/tests/integration/rapidgen"
	"github.com/cosmos/cosmos-sdk/tests/integration/tx/internal"
	gogo_testpb "github.com/cosmos/cosmos-sdk/tests/integration/tx/internal/gogo/testpb"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
)

// TestAminoJSON_Equivalence tests that x/tx/Encoder encoding is equivalent to the legacy Encoder encoding.
// A custom generator is used to generate random messages that are then encoded using both encoders.  The custom
// generator only supports proto.Message (which implement the protoreflect API) so in order to test legacy gogo types
// we end up with a workflow as follows:
//
// 1. Generate a random protobuf proto.Message using the custom generator
// 2. Marshal the proto.Message to protobuf binary bytes
// 3. Unmarshal the protobuf bytes to a gogoproto.Message
// 4. Marshal the gogoproto.Message to amino JSON bytes
// 5. Marshal the proto.Message to amino JSON bytes
// 6. Compare the amino JSON bytes from steps 4 and 5
//
// In order for step 3 to work certain restrictions on the data generated in step 1 must be enforced and are described
// by the mutation of genOpts passed to the generator.
func TestAminoJSON_Equivalence(t *testing.T) {
	fixture := internal.NewSigningFixture(t, internal.SigningFixtureOptions{},
		auth.AppModule{},
		authzmodule.AppModule{},
		bank.AppModule{},
		consensus.AppModule{},
		distribution.AppModule{},
		evidence.AppModule{},
		feegrantmodule.AppModule{},
		gov.AppModule{},
		groupmodule.AppModule{},
		mint.AppModule{},
		slashing.AppModule{},
		staking.AppModule{},
		upgrade.AppModule{},
		vesting.AppModule{},
	)
	aj := aminojson.NewEncoder(aminojson.EncoderOptions{})

	for _, tt := range rapidgen.DefaultGeneratedTypes {
		desc := tt.Pulsar.ProtoReflect().Descriptor()
		name := string(desc.FullName())
		t.Run(name, func(t *testing.T) {
			gen := rapidproto.MessageGenerator(tt.Pulsar, tt.Opts)
			fmt.Printf("testing %s\n", tt.Pulsar.ProtoReflect().Descriptor().FullName())
			rapid.Check(t, func(r *rapid.T) {
				// uncomment to debug; catch a panic and inspect application state
				// defer func() {
				//	if r := recover(); r != nil {
				//		//fmt.Printf("Panic: %+v\n", r)
				//		t.FailNow()
				//	}
				// }()

				msg := gen.Draw(r, "msg")
				postFixPulsarMessage(msg)

				gogo := tt.Gogo
				sanity := tt.Pulsar

				protoBz, err := proto.Marshal(msg)
				require.NoError(r, err)

				err = proto.Unmarshal(protoBz, sanity)
				require.NoError(r, err)

				err = fixture.UnmarshalGogoProto(protoBz, gogo)
				require.NoError(r, err)

				legacyAminoJSON := fixture.MarshalLegacyAminoJSON(t, gogo)
				aminoJSON, err := aj.Marshal(msg)
				require.NoError(r, err)
				if !bytes.Equal(legacyAminoJSON, aminoJSON) {
					require.Failf(r, "JSON mismatch", "legacy: %s\n  x/tx: %s\n",
						string(legacyAminoJSON), string(aminoJSON))
				}

				// test amino json signer handler equivalence
				if !proto.HasExtension(desc.Options(), msgv1.E_Signer) {
					// not signable
					return
				}

				fixture.RequireLegacyAminoEquivalent(t, gogo)
			})
		})
	}
}

// TestAminoJSON_LegacyParity tests that the Encoder encoder produces the same output as the Encoder encoder.
func TestAminoJSON_LegacyParity(t *testing.T) {
	fixture := internal.NewSigningFixture(t, internal.SigningFixtureOptions{},
		auth.AppModule{}, authzmodule.AppModule{},
		bank.AppModule{}, distribution.AppModule{}, slashing.AppModule{}, staking.AppModule{},
		vesting.AppModule{}, gov.AppModule{})
	aj := aminojson.NewEncoder(aminojson.EncoderOptions{})

	addr1 := types.AccAddress("addr1")
	now := time.Now()
	genericAuth, _ := codectypes.NewAnyWithValue(&authztypes.GenericAuthorization{Msg: "foo"})
	pubkeyAny, _ := codectypes.NewAnyWithValue(&secp256k1types.PubKey{Key: []byte("foo")})
	dec5point4 := math.LegacyMustNewDecFromStr("5.4")
	failingBaseAccount := authtypes.NewBaseAccountWithAddress(addr1)
	failingBaseAccount.AccountNumber = stdmath.MaxUint64

	cases := map[string]struct {
		gogo  gogoproto.Message
		fails bool
	}{
		"auth/params": {
			gogo: &authtypes.Params{TxSigLimit: 10},
		},
		"auth/module_account_nil_permissions": {
			gogo: &authtypes.ModuleAccount{
				BaseAccount: authtypes.NewBaseAccountWithAddress(
					addr1,
				),
			},
		},
		"auth/module_account/max_uint64": {
			gogo: &authtypes.ModuleAccount{
				BaseAccount: failingBaseAccount,
			},
			fails: true,
		},
		"auth/module_account_empty_permissions": {
			gogo: &authtypes.ModuleAccount{
				BaseAccount: authtypes.NewBaseAccountWithAddress(
					addr1,
				),
				// empty set and nil are indistinguishable from the protoreflect API since they both
				// marshal to zero proto bytes, there empty set is not supported.
				Permissions: []string{},
			},
			fails: true,
		},
		"auth/base_account": {
			gogo: &authtypes.BaseAccount{Address: addr1.String(), PubKey: pubkeyAny, AccountNumber: 1, Sequence: 2},
		},
		"authz/msg_grant": {
			gogo: &authztypes.MsgGrant{
				Granter: addr1.String(), Grantee: addr1.String(),
				Grant: authztypes.Grant{Expiration: &now, Authorization: genericAuth},
			},
		},
		"authz/msg_update_params": {
			gogo: &authtypes.MsgUpdateParams{Params: authtypes.Params{TxSigLimit: 10}},
		},
		"authz/msg_exec/empty_msgs": {
			gogo: &authztypes.MsgExec{Msgs: []*codectypes.Any{}},
		},
		"distribution/delegator_starting_info": {
			gogo: &disttypes.DelegatorStartingInfo{Stake: math.LegacyNewDec(10)},
		},
		"distribution/delegator_starting_info/non_zero_dec": {
			gogo: &disttypes.DelegatorStartingInfo{Stake: math.LegacyNewDec(10)},
		},
		"distribution/delegation_delegator_reward": {
			gogo: &disttypes.DelegationDelegatorReward{},
		},
		"distribution/msg_withdraw_delegator_reward": {
			gogo: &disttypes.MsgWithdrawDelegatorReward{DelegatorAddress: "foo"},
		},
		"crypto/ed25519": {
			gogo: &ed25519types.PubKey{Key: []byte("key")},
		},
		"crypto/secp256k1": {
			gogo: &secp256k1types.PubKey{Key: []byte("key")},
		},
		"crypto/legacy_amino_pubkey": {
			gogo: &multisig.LegacyAminoPubKey{PubKeys: []*codectypes.Any{pubkeyAny}},
		},
		"crypto/legacy_amino_pubkey_empty": {
			gogo: &multisig.LegacyAminoPubKey{},
		},
		"consensus/evidence_params/duration": {
			gogo: &gov_v1beta1_types.VotingParams{VotingPeriod: 1e9 + 7},
		},
		"consensus/evidence_params/big_duration": {
			gogo: &gov_v1beta1_types.VotingParams{
				VotingPeriod: time.Duration(rapidproto.MaxDurationSeconds*1e9) + 999999999,
			},
		},
		"consensus/evidence_params/too_big_duration": {
			gogo: &gov_v1beta1_types.VotingParams{
				VotingPeriod: time.Duration(rapidproto.MaxDurationSeconds*1e9) + 999999999,
			},
		},
		// amino.dont_omitempty + empty/nil lists produce some surprising results
		"bank/send_authorization/empty_coins": {
			gogo: &banktypes.SendAuthorization{SpendLimit: []types.Coin{}},
		},
		"bank/send_authorization/nil_coins": {
			gogo: &banktypes.SendAuthorization{SpendLimit: nil},
		},
		"bank/send_authorization/empty_list": {
			gogo: &banktypes.SendAuthorization{AllowList: []string{}},
		},
		"bank/send_authorization/nil_list": {
			gogo: &banktypes.SendAuthorization{AllowList: nil},
		},
		"bank/msg_multi_send/nil_everything": {
			gogo: &banktypes.MsgMultiSend{},
		},
		"gov/v1_msg_submit_proposal": {
			gogo: &gov_v1_types.MsgSubmitProposal{},
		},
		"gov/v1_params": {
			gogo: &gov_v1_types.Params{
				Quorum: math.LegacyMustNewDecFromStr("0.33").String(),
			},
		},
		"slashing/params/dec": {
			gogo: &slashingtypes.Params{
				DowntimeJailDuration:    1e9 + 7,
				MinSignedPerWindow:      math.LegacyNewDec(10),
				SlashFractionDoubleSign: math.LegacyZeroDec(),
				SlashFractionDowntime:   math.LegacyZeroDec(),
			},
		},
		"staking/msg_update_params": {
			gogo: &stakingtypes.MsgUpdateParams{
				Params: stakingtypes.Params{
					UnbondingTime:     0,
					KeyRotationFee:    types.Coin{},
					MinCommissionRate: math.LegacyZeroDec(),
				},
			},
		},
		"staking/create_validator": {
			gogo: &stakingtypes.MsgCreateValidator{
				Pubkey: pubkeyAny,
				Commission: stakingtypes.CommissionRates{
					Rate:          dec5point4,
					MaxRate:       math.LegacyZeroDec(),
					MaxChangeRate: math.LegacyZeroDec(),
				},
				MinSelfDelegation: math.NewIntFromUint64(10),
			},
		},
		"staking/msg_cancel_unbonding_delegation_response": {
			gogo: &stakingtypes.MsgCancelUnbondingDelegationResponse{},
		},
		"staking/stake_authorization_empty": {
			gogo: &stakingtypes.StakeAuthorization{},
		},
		"staking/stake_authorization_allow": {
			gogo: &stakingtypes.StakeAuthorization{
				MaxTokens: &types.Coin{Denom: "foo", Amount: math.NewInt(123)},
				Validators: &stakingtypes.StakeAuthorization_AllowList{
					AllowList: &stakingtypes.StakeAuthorization_Validators{
						Address: []string{"foo"},
					},
				},
				AuthorizationType: stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_DELEGATE,
			},
		},
		"staking/stake_authorization_deny": {
			gogo: &stakingtypes.StakeAuthorization{
				MaxTokens: &types.Coin{Denom: "foo", Amount: math.NewInt(123)},
				Validators: &stakingtypes.StakeAuthorization_DenyList{
					DenyList: &stakingtypes.StakeAuthorization_Validators{},
				},
				AuthorizationType: stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_DELEGATE,
			},
		},
		"vesting/base_account_empty": {
			gogo: &vestingtypes.BaseVestingAccount{BaseAccount: &authtypes.BaseAccount{}},
		},
		"vesting/base_account_pubkey": {
			gogo: &vestingtypes.BaseVestingAccount{
				BaseAccount: &authtypes.BaseAccount{PubKey: pubkeyAny},
			},
		},
		"math/int_as_string": {
			gogo: &gogo_testpb.IntAsString{IntAsString: math.NewInt(123)},
		},
		"math/int_as_string/empty": {
			gogo: &gogo_testpb.IntAsString{},
		},
		"math/int_as_bytes": {
			gogo: &gogo_testpb.IntAsBytes{IntAsBytes: math.NewInt(123)},
		},
		"math/int_as_bytes/empty": {
			gogo: &gogo_testpb.IntAsBytes{},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			legacyBytes := fixture.MarshalLegacyAminoJSON(t, tc.gogo)
			dynamicBytes, err := aj.Marshal(fixture.DynamicMessage(t, tc.gogo))
			require.NoError(t, err)

			t.Logf("legacy: %s\n", string(legacyBytes))
			t.Logf("   sut: %s\n", string(dynamicBytes))
			if tc.fails {
				require.NotEqual(t, string(legacyBytes), string(dynamicBytes))
				return
			}
			require.Equal(t, string(legacyBytes), string(dynamicBytes))

			// test amino json signer handler equivalence
			if !proto.HasExtension(fixture.MessageDescriptor(t, tc.gogo).Options(), msgv1.E_Signer) {
				// not signable
				return
			}
			fixture.RequireLegacyAminoEquivalent(t, tc.gogo)
		})
	}
}

func TestSendAuthorization(t *testing.T) {
	encCfg := testutil.MakeTestEncodingConfig(
		codectestutil.CodecOptions{},
		auth.AppModule{},
		authzmodule.AppModule{},
		distribution.AppModule{},
		bank.AppModule{},
	)

	aj := aminojson.NewEncoder(aminojson.EncoderOptions{})

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
	require.NoError(t, err)

	// !!!
	//  empty []string is not the same as nil []string.  this is a bug in gogo.
	// `[]string` -> proto.Marshal -> legacyAmino.UnmarshalProto (unmarshals empty slice as nil)
	//    -> legacyAmino.MarshalJson -> `null`
	// `[]string` -> [proto.Marshal -> pulsar.Unmarshal] -> amino.MarshalJson -> `[]`
	require.Nil(t, gogo.SpendLimit)
	require.Nil(t, sanityPulsar.SpendLimit)
	require.NotNil(t, pulsar.SpendLimit)
	require.Zero(t, len(pulsar.SpendLimit))

	legacyAminoJSON, err := encCfg.Amino.MarshalJSON(gogo)
	require.NoError(t, err)
	aminoJSON, err := aj.Marshal(sanityPulsar)
	require.NoError(t, err)

	require.Equal(t, string(legacyAminoJSON), string(aminoJSON))

	aminoJSON, err = aj.Marshal(pulsar)
	require.NoError(t, err)

	// at this point, pulsar.SpendLimit = [], and gogo.SpendLimit = nil, but they will both marshal to `[]`
	// this is *only* possible because of Cosmos SDK's custom MarshalJSON method for Coins
	require.Equal(t, string(legacyAminoJSON), string(aminoJSON))
}

func TestDecimalMutation(t *testing.T) {
	encCfg := testutil.MakeTestEncodingConfig(codectestutil.CodecOptions{}, staking.AppModule{})
	rates := &stakingtypes.CommissionRates{}
	rateBz, _ := encCfg.Amino.MarshalJSON(rates)
	require.Equal(t, `{"rate":"0","max_rate":"0","max_change_rate":"0"}`, string(rateBz))
	_, err := gogoproto.Marshal(rates)
	require.NoError(t, err)
	rateBz, _ = encCfg.Amino.MarshalJSON(rates)

	// prior to the merge of https://github.com/cosmos/cosmos-sdk/pull/15506
	// gogoproto.Marshal would mutate Decimal fields changing JSON output as shown in the assertions below
	// require.NotEqual(t, `{"rate":"0","max_rate":"0","max_change_rate":"0"}`, string(rateBz))
	// require.Equal(t,
	//	`{"rate":"0.000000000000000000","max_rate":"0.000000000000000000","max_change_rate":"0.000000000000000000"}`,
	//	string(rateBz))

	// This is no longer the case, new behavior:
	require.Equal(t, `{"rate":"0","max_rate":"0","max_change_rate":"0"}`, string(rateBz))
}

func postFixPulsarMessage(msg proto.Message) {
	if m, ok := msg.(*authapi.ModuleAccount); ok {
		if m.BaseAccount == nil {
			m.BaseAccount = &authapi.BaseAccount{}
		}
		_, _, bz := testdata.KeyTestPubAddr()
		// always set address to a valid bech32 address
		text, _ := bech32.ConvertAndEncode("cosmos", bz)
		m.BaseAccount.Address = text

		// see negative test
		if len(m.Permissions) == 0 {
			m.Permissions = nil
		}
	}
}
