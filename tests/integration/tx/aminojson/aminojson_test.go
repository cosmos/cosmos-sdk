package aminojson

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cosmos/cosmos-proto/rapidproto"
	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"pgregory.net/rapid"

	authapi "cosmossdk.io/api/cosmos/auth/v1beta1"
	authzapi "cosmossdk.io/api/cosmos/authz/v1beta1"
	bankapi "cosmossdk.io/api/cosmos/bank/v1beta1"
	v1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
	"cosmossdk.io/api/cosmos/crypto/ed25519"
	multisigapi "cosmossdk.io/api/cosmos/crypto/multisig"
	"cosmossdk.io/api/cosmos/crypto/secp256k1"
	distapi "cosmossdk.io/api/cosmos/distribution/v1beta1"
	gov_v1_api "cosmossdk.io/api/cosmos/gov/v1"
	gov_v1beta1_api "cosmossdk.io/api/cosmos/gov/v1beta1"
	msgv1 "cosmossdk.io/api/cosmos/msg/v1"
	slashingapi "cosmossdk.io/api/cosmos/slashing/v1beta1"
	stakingapi "cosmossdk.io/api/cosmos/staking/v1beta1"
	txv1beta1 "cosmossdk.io/api/cosmos/tx/v1beta1"
	vestingapi "cosmossdk.io/api/cosmos/vesting/v1beta1"
	"cosmossdk.io/math"
	"cosmossdk.io/x/auth"
	"cosmossdk.io/x/auth/migrations/legacytx"
	"cosmossdk.io/x/auth/signing"
	"cosmossdk.io/x/auth/tx"
	authtypes "cosmossdk.io/x/auth/types"
	"cosmossdk.io/x/auth/vesting"
	vestingtypes "cosmossdk.io/x/auth/vesting/types"
	authztypes "cosmossdk.io/x/authz"
	authzmodule "cosmossdk.io/x/authz/module"
	"cosmossdk.io/x/bank"
	banktypes "cosmossdk.io/x/bank/types"
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
	signing_testutil "cosmossdk.io/x/tx/signing/testutil"
	"cosmossdk.io/x/upgrade"

	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	ed25519types "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	secp256k1types "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/tests/integration/rapidgen"
	gogo_testpb "github.com/cosmos/cosmos-sdk/tests/integration/tx/internal/gogo/testpb"
	pulsar_testpb "github.com/cosmos/cosmos-sdk/tests/integration/tx/internal/pulsar/testpb"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/consensus"
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
	encCfg := testutil.MakeTestEncodingConfig(
		codectestutil.CodecOptions{}, auth.AppModule{}, authzmodule.AppModule{}, bank.AppModule{},
		consensus.AppModule{}, distribution.AppModule{}, evidence.AppModule{}, feegrantmodule.AppModule{},
		gov.AppModule{}, groupmodule.AppModule{}, mint.AppModule{},
		slashing.AppModule{}, staking.AppModule{}, upgrade.AppModule{}, vesting.AppModule{})
	legacytx.RegressionTestingAminoCodec = encCfg.Amino
	aj := aminojson.NewEncoder(aminojson.EncoderOptions{DoNotSortFields: true})

	for _, tt := range rapidgen.DefaultGeneratedTypes {
		desc := tt.Pulsar.ProtoReflect().Descriptor()
		name := string(desc.FullName())
		t.Run(name, func(t *testing.T) {
			gen := rapidproto.MessageGenerator(tt.Pulsar, tt.Opts)
			fmt.Printf("testing %s\n", tt.Pulsar.ProtoReflect().Descriptor().FullName())
			rapid.Check(t, func(t *rapid.T) {
				// uncomment to debug; catch a panic and inspect application state
				// defer func() {
				//	if r := recover(); r != nil {
				//		//fmt.Printf("Panic: %+v\n", r)
				//		t.FailNow()
				//	}
				// }()

				msg := gen.Draw(t, "msg")
				postFixPulsarMessage(msg)
				// txBuilder.GetTx will fail if the msg has no signers
				// so it does not make sense to run these cases, apparently.
				signers, err := encCfg.TxConfig.SigningContext().GetSigners(msg)
				if len(signers) == 0 {
					// skip
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

				legacyAminoJSON, err := encCfg.Amino.MarshalJSON(gogo)
				require.NoError(t, err)
				aminoJSON, err := aj.Marshal(msg)
				require.NoError(t, err)
				require.Equal(t, string(legacyAminoJSON), string(aminoJSON))

				// test amino json signer handler equivalence
				if !proto.HasExtension(desc.Options(), msgv1.E_Signer) {
					// not signable
					return
				}

				handlerOptions := signing_testutil.HandlerArgumentOptions{
					ChainID:       "test-chain",
					Memo:          "sometestmemo",
					Msg:           tt.Pulsar,
					AccNum:        1,
					AccSeq:        2,
					SignerAddress: "signerAddress",
					Fee: &txv1beta1.Fee{
						Amount: []*v1beta1.Coin{{Denom: "uatom", Amount: "1000"}},
					},
				}

				signerData, txData, err := signing_testutil.MakeHandlerArguments(handlerOptions)
				require.NoError(t, err)

				handler := aminojson.NewSignModeHandler(aminojson.SignModeHandlerOptions{})
				signBz, err := handler.GetSignBytes(context.Background(), signerData, txData)
				require.NoError(t, err)

				legacyHandler := tx.NewSignModeLegacyAminoJSONHandler()
				txBuilder := encCfg.TxConfig.NewTxBuilder()
				require.NoError(t, txBuilder.SetMsgs([]types.Msg{tt.Gogo}...))
				txBuilder.SetMemo(handlerOptions.Memo)
				txBuilder.SetFeeAmount(types.Coins{types.NewInt64Coin("uatom", 1000)})
				theTx := txBuilder.GetTx()

				legacySigningData := signing.SignerData{
					ChainID:       handlerOptions.ChainID,
					Address:       handlerOptions.SignerAddress,
					AccountNumber: handlerOptions.AccNum,
					Sequence:      handlerOptions.AccSeq,
				}
				legacySignBz, err := legacyHandler.GetSignBytes(signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
					legacySigningData, theTx)
				require.NoError(t, err)
				require.Equal(t, string(legacySignBz), string(signBz))
			})
		})
	}
}

func newAny(t *testing.T, msg proto.Message) *anypb.Any {
	t.Helper()
	bz, err := proto.Marshal(msg)
	require.NoError(t, err)
	typeName := fmt.Sprintf("/%s", msg.ProtoReflect().Descriptor().FullName())
	return &anypb.Any{
		TypeUrl: typeName,
		Value:   bz,
	}
}

// TestAminoJSON_LegacyParity tests that the Encoder encoder produces the same output as the Encoder encoder.
func TestAminoJSON_LegacyParity(t *testing.T) {
	encCfg := testutil.MakeTestEncodingConfig(codectestutil.CodecOptions{}, auth.AppModule{}, authzmodule.AppModule{},
		bank.AppModule{}, distribution.AppModule{}, slashing.AppModule{}, staking.AppModule{},
		vesting.AppModule{}, gov.AppModule{})
	legacytx.RegressionTestingAminoCodec = encCfg.Amino

	aj := aminojson.NewEncoder(aminojson.EncoderOptions{DoNotSortFields: true})
	addr1 := types.AccAddress("addr1")
	now := time.Now()

	genericAuth, _ := codectypes.NewAnyWithValue(&authztypes.GenericAuthorization{Msg: "foo"})
	genericAuthPulsar := newAny(t, &authzapi.GenericAuthorization{Msg: "foo"})
	pubkeyAny, _ := codectypes.NewAnyWithValue(&secp256k1types.PubKey{Key: []byte("foo")})
	pubkeyAnyPulsar := newAny(t, &secp256k1.PubKey{Key: []byte("foo")})
	dec10bz, _ := math.LegacyNewDec(10).Marshal()
	int123bz, _ := math.NewInt(123).Marshal()

	cases := map[string]struct {
		gogo               gogoproto.Message
		pulsar             proto.Message
		pulsarMarshalFails bool

		// this will fail in cases where a lossy encoding of an empty array to protobuf occurs. the unmarshalled bytes
		// represent the array as nil, and a subsequent marshal to JSON represent the array as null instead of empty.
		roundTripUnequal bool

		// pulsar does not support marshaling a math.Dec as anything except a string.  Therefore, we cannot unmarshal
		// a pulsar encoded Math.dec (the string representation of a Decimal) into a gogo Math.dec (expecting an int64).
		protoUnmarshalFails bool
	}{
		"auth/params": {gogo: &authtypes.Params{TxSigLimit: 10}, pulsar: &authapi.Params{TxSigLimit: 10}},
		"auth/module_account": {
			gogo: &authtypes.ModuleAccount{
				BaseAccount: authtypes.NewBaseAccountWithAddress(addr1), Permissions: []string{},
			},
			pulsar: &authapi.ModuleAccount{
				BaseAccount: &authapi.BaseAccount{Address: addr1.String()}, Permissions: []string{},
			},
			roundTripUnequal: true,
		},
		"auth/base_account": {
			gogo:   &authtypes.BaseAccount{Address: addr1.String(), PubKey: pubkeyAny},
			pulsar: &authapi.BaseAccount{Address: addr1.String(), PubKey: pubkeyAnyPulsar},
		},
		"authz/msg_grant": {
			gogo: &authztypes.MsgGrant{
				Grant: authztypes.Grant{Expiration: &now, Authorization: genericAuth},
			},
			pulsar: &authzapi.MsgGrant{
				Grant: &authzapi.Grant{Expiration: timestamppb.New(now), Authorization: genericAuthPulsar},
			},
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
			gogo:                &disttypes.DelegatorStartingInfo{Stake: math.LegacyNewDec(10)},
			pulsar:              &distapi.DelegatorStartingInfo{Stake: "10.000000000000000000"},
			protoUnmarshalFails: true,
		},
		"distribution/delegation_delegator_reward": {
			gogo:   &disttypes.DelegationDelegatorReward{},
			pulsar: &distapi.DelegationDelegatorReward{},
		},
		"distribution/msg_withdraw_delegator_reward": {
			gogo:   &disttypes.MsgWithdrawDelegatorReward{DelegatorAddress: "foo"},
			pulsar: &distapi.MsgWithdrawDelegatorReward{DelegatorAddress: "foo"},
		},
		"crypto/ed25519": {
			gogo:   &ed25519types.PubKey{Key: []byte("key")},
			pulsar: &ed25519.PubKey{Key: []byte("key")},
		},
		"crypto/secp256k1": {
			gogo:   &secp256k1types.PubKey{Key: []byte("key")},
			pulsar: &secp256k1.PubKey{Key: []byte("key")},
		},
		"crypto/legacy_amino_pubkey": {
			gogo:   &multisig.LegacyAminoPubKey{PubKeys: []*codectypes.Any{pubkeyAny}},
			pulsar: &multisigapi.LegacyAminoPubKey{PublicKeys: []*anypb.Any{pubkeyAnyPulsar}},
		},
		"crypto/legacy_amino_pubkey/empty": {
			gogo:   &multisig.LegacyAminoPubKey{},
			pulsar: &multisigapi.LegacyAminoPubKey{},
		},
		"consensus/evidence_params/duration": {
			gogo:   &gov_v1beta1_types.VotingParams{VotingPeriod: 1e9 + 7},
			pulsar: &gov_v1beta1_api.VotingParams{VotingPeriod: &durationpb.Duration{Seconds: 1, Nanos: 7}},
		},
		"consensus/evidence_params/big_duration": {
			gogo: &gov_v1beta1_types.VotingParams{VotingPeriod: time.Duration(rapidproto.MaxDurationSeconds*1e9) + 999999999},
			pulsar: &gov_v1beta1_api.VotingParams{VotingPeriod: &durationpb.Duration{
				Seconds: rapidproto.MaxDurationSeconds, Nanos: 999999999,
			}},
		},
		"consensus/evidence_params/too_big_duration": {
			gogo: &gov_v1beta1_types.VotingParams{VotingPeriod: time.Duration(rapidproto.MaxDurationSeconds*1e9) + 999999999},
			pulsar: &gov_v1beta1_api.VotingParams{VotingPeriod: &durationpb.Duration{
				Seconds: rapidproto.MaxDurationSeconds + 1, Nanos: 999999999,
			}},
			pulsarMarshalFails: true,
		},
		// amino.dont_omitempty + empty/nil lists produce some surprising results
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
		"gov/v1_msg_submit_proposal": {
			gogo:   &gov_v1_types.MsgSubmitProposal{},
			pulsar: &gov_v1_api.MsgSubmitProposal{},
		},
		"slashing/params/empty_dec": {
			gogo:   &slashingtypes.Params{DowntimeJailDuration: 1e9 + 7},
			pulsar: &slashingapi.Params{DowntimeJailDuration: &durationpb.Duration{Seconds: 1, Nanos: 7}},
		},
		// This test cases demonstrates the expected contract and proper way to set a cosmos.Dec field represented
		// as bytes in protobuf message, namely:
		// dec10bz, _ := types.NewDec(10).Marshal()
		"slashing/params/dec": {
			gogo: &slashingtypes.Params{
				DowntimeJailDuration: 1e9 + 7,
				MinSignedPerWindow:   math.LegacyNewDec(10),
			},
			pulsar: &slashingapi.Params{
				DowntimeJailDuration: &durationpb.Duration{Seconds: 1, Nanos: 7},
				MinSignedPerWindow:   dec10bz,
			},
		},
		"staking/msg_update_params": {
			gogo: &stakingtypes.MsgUpdateParams{
				Params: stakingtypes.Params{
					UnbondingTime:  0,
					KeyRotationFee: types.Coin{},
				},
			},
			pulsar: &stakingapi.MsgUpdateParams{
				Params: &stakingapi.Params{
					UnbondingTime:  &durationpb.Duration{Seconds: 0},
					KeyRotationFee: &v1beta1.Coin{},
				},
			},
		},
		"staking/create_validator": {
			gogo: &stakingtypes.MsgCreateValidator{Pubkey: pubkeyAny},
			pulsar: &stakingapi.MsgCreateValidator{
				Pubkey:      pubkeyAnyPulsar,
				Description: &stakingapi.Description{},
				Commission:  &stakingapi.CommissionRates{},
				Value:       &v1beta1.Coin{},
			},
		},
		"staking/msg_cancel_unbonding_delegation_response": {
			gogo:   &stakingtypes.MsgCancelUnbondingDelegationResponse{},
			pulsar: &stakingapi.MsgCancelUnbondingDelegationResponse{},
		},
		"staking/stake_authorization_empty": {
			gogo:   &stakingtypes.StakeAuthorization{},
			pulsar: &stakingapi.StakeAuthorization{},
		},
		"staking/stake_authorization_allow": {
			gogo: &stakingtypes.StakeAuthorization{
				Validators: &stakingtypes.StakeAuthorization_AllowList{
					AllowList: &stakingtypes.StakeAuthorization_Validators{Address: []string{"foo"}},
				},
			},
			pulsar: &stakingapi.StakeAuthorization{
				Validators: &stakingapi.StakeAuthorization_AllowList{
					AllowList: &stakingapi.StakeAuthorization_Validators{Address: []string{"foo"}},
				},
			},
		},
		"vesting/base_account_empty": {
			gogo:   &vestingtypes.BaseVestingAccount{BaseAccount: &authtypes.BaseAccount{}},
			pulsar: &vestingapi.BaseVestingAccount{BaseAccount: &authapi.BaseAccount{}},
		},
		"vesting/base_account_pubkey": {
			gogo:   &vestingtypes.BaseVestingAccount{BaseAccount: &authtypes.BaseAccount{PubKey: pubkeyAny}},
			pulsar: &vestingapi.BaseVestingAccount{BaseAccount: &authapi.BaseAccount{PubKey: pubkeyAnyPulsar}},
		},
		"math/int_as_string": {
			gogo:   &gogo_testpb.IntAsString{IntAsString: math.NewInt(123)},
			pulsar: &pulsar_testpb.IntAsString{IntAsString: "123"},
		},
		"math/int_as_string/empty": {
			gogo:   &gogo_testpb.IntAsString{},
			pulsar: &pulsar_testpb.IntAsString{},
		},
		"math/int_as_bytes": {
			gogo:   &gogo_testpb.IntAsBytes{IntAsBytes: math.NewInt(123)},
			pulsar: &pulsar_testpb.IntAsBytes{IntAsBytes: int123bz},
		},
		"math/int_as_bytes/empty": {
			gogo:   &gogo_testpb.IntAsBytes{},
			pulsar: &pulsar_testpb.IntAsBytes{},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			gogoBytes, err := encCfg.Amino.MarshalJSON(tc.gogo)
			require.NoError(t, err)

			pulsarBytes, err := aj.Marshal(tc.pulsar)
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
			require.NoError(t, err)
			if tc.roundTripUnequal {
				require.NotEqual(t, string(gogoBytes), string(newGogoBytes))
				return
			}
			require.Equal(t, string(gogoBytes), string(newGogoBytes))

			// test amino json signer handler equivalence
			msg, ok := tc.gogo.(legacytx.LegacyMsg)
			if !ok {
				// not signable
				return
			}

			handlerOptions := signing_testutil.HandlerArgumentOptions{
				ChainID:       "test-chain",
				Memo:          "sometestmemo",
				Msg:           tc.pulsar,
				AccNum:        1,
				AccSeq:        2,
				SignerAddress: "signerAddress",
				Fee: &txv1beta1.Fee{
					Amount: []*v1beta1.Coin{{Denom: "uatom", Amount: "1000"}},
				},
			}

			signerData, txData, err := signing_testutil.MakeHandlerArguments(handlerOptions)
			require.NoError(t, err)

			handler := aminojson.NewSignModeHandler(aminojson.SignModeHandlerOptions{})
			signBz, err := handler.GetSignBytes(context.Background(), signerData, txData)
			require.NoError(t, err)

			legacyHandler := tx.NewSignModeLegacyAminoJSONHandler()
			txBuilder := encCfg.TxConfig.NewTxBuilder()
			require.NoError(t, txBuilder.SetMsgs([]types.Msg{msg}...))
			txBuilder.SetMemo(handlerOptions.Memo)
			txBuilder.SetFeeAmount(types.Coins{types.NewInt64Coin("uatom", 1000)})
			theTx := txBuilder.GetTx()

			legacySigningData := signing.SignerData{
				ChainID:       handlerOptions.ChainID,
				Address:       handlerOptions.SignerAddress,
				AccountNumber: handlerOptions.AccNum,
				Sequence:      handlerOptions.AccSeq,
			}
			legacySignBz, err := legacyHandler.GetSignBytes(signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
				legacySigningData, theTx)
			require.NoError(t, err)
			require.Equal(t, string(legacySignBz), string(signBz))
		})
	}
}

func TestSendAuthorization(t *testing.T) {
	encCfg := testutil.MakeTestEncodingConfig(codectestutil.CodecOptions{}, auth.AppModule{}, authzmodule.AppModule{},
		distribution.AppModule{}, bank.AppModule{})

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
