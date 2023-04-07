package aminojson

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	cosmos_proto "github.com/cosmos/cosmos-proto"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"pgregory.net/rapid"

	"github.com/cosmos/cosmos-proto/rapidproto"
	gogoproto "github.com/cosmos/gogoproto/proto"

	"cosmossdk.io/api/amino"
	authapi "cosmossdk.io/api/cosmos/auth/v1beta1"
	authzapi "cosmossdk.io/api/cosmos/authz/v1beta1"
	bankapi "cosmossdk.io/api/cosmos/bank/v1beta1"
	v1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
	consensusapi "cosmossdk.io/api/cosmos/consensus/v1"
	"cosmossdk.io/api/cosmos/crypto/ed25519"
	multisigapi "cosmossdk.io/api/cosmos/crypto/multisig"
	"cosmossdk.io/api/cosmos/crypto/secp256k1"
	distapi "cosmossdk.io/api/cosmos/distribution/v1beta1"
	evidenceapi "cosmossdk.io/api/cosmos/evidence/v1beta1"
	feegrantapi "cosmossdk.io/api/cosmos/feegrant/v1beta1"
	gov_v1_api "cosmossdk.io/api/cosmos/gov/v1"
	gov_v1beta1_api "cosmossdk.io/api/cosmos/gov/v1beta1"
	groupapi "cosmossdk.io/api/cosmos/group/v1"
	mintapi "cosmossdk.io/api/cosmos/mint/v1beta1"
	paramsapi "cosmossdk.io/api/cosmos/params/v1beta1"
	slashingapi "cosmossdk.io/api/cosmos/slashing/v1beta1"
	stakingapi "cosmossdk.io/api/cosmos/staking/v1beta1"
	txv1beta1 "cosmossdk.io/api/cosmos/tx/v1beta1"
	upgradeapi "cosmossdk.io/api/cosmos/upgrade/v1beta1"
	vestingapi "cosmossdk.io/api/cosmos/vesting/v1beta1"
	"cosmossdk.io/x/evidence"
	evidencetypes "cosmossdk.io/x/evidence/types"
	feegranttypes "cosmossdk.io/x/feegrant"
	feegrantmodule "cosmossdk.io/x/feegrant/module"
	"cosmossdk.io/x/tx/signing/aminojson"
	signing_testutil "cosmossdk.io/x/tx/signing/testutil"
	"cosmossdk.io/x/upgrade"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	ed25519types "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	secp256k1types "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	gogo_testpb "github.com/cosmos/cosmos-sdk/tests/integration/aminojson/internal/gogo/testpb"
	pulsar_testpb "github.com/cosmos/cosmos-sdk/tests/integration/aminojson/internal/pulsar/testpb"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	authztypes "github.com/cosmos/cosmos-sdk/x/authz"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/consensus"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	gov_v1_types "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	gov_v1beta1_types "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	grouptypes "github.com/cosmos/cosmos-sdk/x/group"
	groupmodule "github.com/cosmos/cosmos-sdk/x/group/module"
	"github.com/cosmos/cosmos-sdk/x/mint"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
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

func withDecisionPolicy(opts rapidproto.GeneratorOptions) rapidproto.GeneratorOptions {
	return opts.
		WithAnyTypes(
			&groupapi.ThresholdDecisionPolicy{},
			&groupapi.PercentageDecisionPolicy{}).
		WithDisallowNil().
		WithInterfaceHint("cosmos.group.v1.DecisionPolicy", &groupapi.ThresholdDecisionPolicy{}).
		WithInterfaceHint("cosmos.group.v1.DecisionPolicy", &groupapi.PercentageDecisionPolicy{})
}

func generatorFieldMapper(t *rapid.T, field protoreflect.FieldDescriptor, name string) (protoreflect.Value, bool) {
	opts := field.Options()
	switch {
	case proto.HasExtension(opts, cosmos_proto.E_Scalar):
		scalar := proto.GetExtension(opts, cosmos_proto.E_Scalar).(string)
		switch scalar {
		case "cosmos.Int":
			i32 := rapid.Int32().Draw(t, name)
			return protoreflect.ValueOfString(fmt.Sprintf("%d", i32)), true
		case "cosmos.Dec":
			return protoreflect.ValueOfString(""), true
		}
	case field.Kind() == protoreflect.BytesKind:
		if proto.HasExtension(opts, amino.E_Encoding) {
			encoding := proto.GetExtension(opts, amino.E_Encoding).(string)
			if encoding == "cosmos_dec_bytes" {
				return protoreflect.ValueOfBytes([]byte{}), true
			}
		}
	}

	return protoreflect.Value{}, false
}

var (
	genOpts = rapidproto.GeneratorOptions{
		Resolver:  protoregistry.GlobalTypes,
		FieldMaps: []rapidproto.FieldMapper{generatorFieldMapper},
	}
	genTypes = []generatedType{
		// auth
		genType(&authtypes.Params{}, &authapi.Params{}, genOpts),
		genType(&authtypes.BaseAccount{}, &authapi.BaseAccount{}, genOpts.WithAnyTypes(&ed25519.PubKey{})),
		genType(&authtypes.ModuleAccount{}, &authapi.ModuleAccount{}, genOpts.WithAnyTypes(&ed25519.PubKey{})),
		genType(&authtypes.ModuleCredential{}, &authapi.ModuleCredential{}, genOpts),
		genType(&authtypes.MsgUpdateParams{}, &authapi.MsgUpdateParams{}, genOpts.WithDisallowNil()),

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

		// crypto
		genType(&multisig.LegacyAminoPubKey{}, &multisigapi.LegacyAminoPubKey{},
			genOpts.WithAnyTypes(&ed25519.PubKey{}, &secp256k1.PubKey{})),

		// distribution
		genType(&disttypes.MsgWithdrawDelegatorReward{}, &distapi.MsgWithdrawDelegatorReward{}, genOpts),
		genType(&disttypes.MsgWithdrawValidatorCommission{}, &distapi.MsgWithdrawValidatorCommission{}, genOpts),
		genType(&disttypes.MsgSetWithdrawAddress{}, &distapi.MsgSetWithdrawAddress{}, genOpts),
		genType(&disttypes.MsgFundCommunityPool{}, &distapi.MsgFundCommunityPool{}, genOpts),
		genType(&disttypes.MsgUpdateParams{}, &distapi.MsgUpdateParams{}, genOpts.WithDisallowNil()),
		genType(&disttypes.MsgCommunityPoolSpend{}, &distapi.MsgCommunityPoolSpend{}, genOpts),
		genType(&disttypes.MsgDepositValidatorRewardsPool{}, &distapi.MsgDepositValidatorRewardsPool{}, genOpts),
		genType(&disttypes.Params{}, &distapi.Params{}, genOpts),

		// evidence
		genType(&evidencetypes.Equivocation{}, &evidenceapi.Equivocation{}, genOpts.WithDisallowNil()),
		genType(&evidencetypes.MsgSubmitEvidence{}, &evidenceapi.MsgSubmitEvidence{},
			genOpts.WithAnyTypes(&evidenceapi.Equivocation{}).
				WithDisallowNil().
				WithInterfaceHint("cosmos.evidence.v1beta1.Evidence", &evidenceapi.Equivocation{})),

		// feegrant
		genType(&feegranttypes.MsgGrantAllowance{}, &feegrantapi.MsgGrantAllowance{},
			genOpts.WithDisallowNil().
				WithAnyTypes(
					&feegrantapi.BasicAllowance{},
					&feegrantapi.PeriodicAllowance{}).
				WithInterfaceHint("cosmos.feegrant.v1beta1.FeeAllowanceI", &feegrantapi.BasicAllowance{}).
				WithInterfaceHint("cosmos.feegrant.v1beta1.FeeAllowanceI", &feegrantapi.PeriodicAllowance{}),
		),
		genType(&feegranttypes.MsgRevokeAllowance{}, &feegrantapi.MsgRevokeAllowance{}, genOpts),
		genType(&feegranttypes.BasicAllowance{}, &feegrantapi.BasicAllowance{}, genOpts.WithDisallowNil()),
		genType(&feegranttypes.PeriodicAllowance{}, &feegrantapi.PeriodicAllowance{}, genOpts.WithDisallowNil()),
		genType(&feegranttypes.AllowedMsgAllowance{}, &feegrantapi.AllowedMsgAllowance{},
			genOpts.WithDisallowNil().
				WithAnyTypes(
					&feegrantapi.BasicAllowance{},
					&feegrantapi.PeriodicAllowance{}).
				WithInterfaceHint("cosmos.feegrant.v1beta1.FeeAllowanceI", &feegrantapi.BasicAllowance{}).
				WithInterfaceHint("cosmos.feegrant.v1beta1.FeeAllowanceI", &feegrantapi.PeriodicAllowance{}),
		),

		// gov v1beta1
		genType(&gov_v1beta1_types.MsgSubmitProposal{}, &gov_v1beta1_api.MsgSubmitProposal{},
			genOpts.WithAnyTypes(&gov_v1beta1_api.TextProposal{}).
				WithDisallowNil().
				WithInterfaceHint("cosmos.gov.v1beta1.Content", &gov_v1beta1_api.TextProposal{}),
		),
		genType(&gov_v1beta1_types.MsgDeposit{}, &gov_v1beta1_api.MsgDeposit{}, genOpts),
		genType(&gov_v1beta1_types.MsgVote{}, &gov_v1beta1_api.MsgVote{}, genOpts),
		genType(&gov_v1beta1_types.MsgVoteWeighted{}, &gov_v1beta1_api.MsgVoteWeighted{}, genOpts),
		genType(&gov_v1beta1_types.TextProposal{}, &gov_v1beta1_api.TextProposal{}, genOpts),

		// gov v1
		genType(&gov_v1_types.MsgSubmitProposal{}, &gov_v1_api.MsgSubmitProposal{},
			genOpts.WithAnyTypes(&gov_v1_api.MsgVote{}, &gov_v1_api.MsgVoteWeighted{}, &gov_v1_api.MsgDeposit{},
				&gov_v1_api.MsgExecLegacyContent{}, &gov_v1_api.MsgUpdateParams{}).
				WithInterfaceHint("cosmos.gov.v1beta1.Content", &gov_v1beta1_api.TextProposal{}).
				WithDisallowNil(),
		),
		genType(&gov_v1_types.MsgDeposit{}, &gov_v1_api.MsgDeposit{}, genOpts),
		genType(&gov_v1_types.MsgVote{}, &gov_v1_api.MsgVote{}, genOpts),
		genType(&gov_v1_types.MsgVoteWeighted{}, &gov_v1_api.MsgVoteWeighted{}, genOpts),
		genType(&gov_v1_types.MsgExecLegacyContent{}, &gov_v1_api.MsgExecLegacyContent{},
			genOpts.WithAnyTypes(&gov_v1beta1_api.TextProposal{}).
				WithDisallowNil().
				WithInterfaceHint("cosmos.gov.v1beta1.Content", &gov_v1beta1_api.TextProposal{})),
		genType(&gov_v1_types.MsgUpdateParams{}, &gov_v1_api.MsgUpdateParams{}, genOpts.WithDisallowNil()),

		// group
		genType(&grouptypes.MsgCreateGroup{}, &groupapi.MsgCreateGroup{}, genOpts),
		genType(&grouptypes.MsgUpdateGroupMembers{}, &groupapi.MsgUpdateGroupMembers{}, genOpts),
		genType(&grouptypes.MsgUpdateGroupAdmin{}, &groupapi.MsgUpdateGroupAdmin{}, genOpts),
		genType(&grouptypes.MsgUpdateGroupMetadata{}, &groupapi.MsgUpdateGroupMetadata{}, genOpts),
		genType(&grouptypes.MsgCreateGroupWithPolicy{}, &groupapi.MsgCreateGroupWithPolicy{},
			withDecisionPolicy(genOpts)),
		genType(&grouptypes.MsgCreateGroupPolicy{}, &groupapi.MsgCreateGroupPolicy{},
			withDecisionPolicy(genOpts)),
		genType(&grouptypes.MsgUpdateGroupPolicyAdmin{}, &groupapi.MsgUpdateGroupPolicyAdmin{}, genOpts),
		genType(&grouptypes.MsgUpdateGroupPolicyDecisionPolicy{}, &groupapi.MsgUpdateGroupPolicyDecisionPolicy{},
			withDecisionPolicy(genOpts)),
		genType(&grouptypes.MsgUpdateGroupPolicyMetadata{}, &groupapi.MsgUpdateGroupPolicyMetadata{}, genOpts),
		genType(&grouptypes.MsgSubmitProposal{}, &groupapi.MsgSubmitProposal{},
			genOpts.WithDisallowNil().
				WithAnyTypes(&groupapi.MsgCreateGroup{}, &groupapi.MsgUpdateGroupMembers{}).
				WithInterfaceHint("cosmos.base.v1beta1.Msg", &groupapi.MsgCreateGroup{}).
				WithInterfaceHint("cosmos.base.v1beta1.Msg", &groupapi.MsgUpdateGroupMembers{}),
		),
		genType(&grouptypes.MsgVote{}, &groupapi.MsgVote{}, genOpts),
		genType(&grouptypes.MsgExec{}, &groupapi.MsgExec{}, genOpts),
		genType(&grouptypes.MsgLeaveGroup{}, &groupapi.MsgLeaveGroup{}, genOpts),

		// mint
		genType(&minttypes.Params{}, &mintapi.Params{}, genOpts),
		genType(&minttypes.MsgUpdateParams{}, &mintapi.MsgUpdateParams{}, genOpts.WithDisallowNil()),

		// params
		genType(&proposal.ParameterChangeProposal{}, &paramsapi.ParameterChangeProposal{}, genOpts),

		// slashing
		genType(&slashingtypes.Params{}, &slashingapi.Params{}, genOpts.WithDisallowNil()),
		genType(&slashingtypes.MsgUnjail{}, &slashingapi.MsgUnjail{}, genOpts),
		genType(&slashingtypes.MsgUpdateParams{}, &slashingapi.MsgUpdateParams{}, genOpts.WithDisallowNil()),

		// staking
		genType(&stakingtypes.MsgCreateValidator{}, &stakingapi.MsgCreateValidator{},
			genOpts.WithDisallowNil().
				WithAnyTypes(&ed25519.PubKey{}).
				WithInterfaceHint("cosmos.crypto.PubKey", &ed25519.PubKey{}),
		),
		genType(&stakingtypes.MsgEditValidator{}, &stakingapi.MsgEditValidator{}, genOpts.WithDisallowNil()),
		genType(&stakingtypes.MsgDelegate{}, &stakingapi.MsgDelegate{}, genOpts.WithDisallowNil()),
		genType(&stakingtypes.MsgUndelegate{}, &stakingapi.MsgUndelegate{}, genOpts.WithDisallowNil()),
		genType(&stakingtypes.MsgBeginRedelegate{}, &stakingapi.MsgBeginRedelegate{}, genOpts.WithDisallowNil()),
		genType(&stakingtypes.MsgUpdateParams{}, &stakingapi.MsgUpdateParams{}, genOpts.WithDisallowNil()),
		genType(&stakingtypes.StakeAuthorization{}, &stakingapi.StakeAuthorization{}, genOpts),

		// upgrade
		genType(&upgradetypes.CancelSoftwareUpgradeProposal{}, &upgradeapi.CancelSoftwareUpgradeProposal{}, genOpts),       // nolint:staticcheck // testing legacy code path
		genType(&upgradetypes.SoftwareUpgradeProposal{}, &upgradeapi.SoftwareUpgradeProposal{}, genOpts.WithDisallowNil()), // nolint:staticcheck // testing legacy code path
		genType(&upgradetypes.Plan{}, &upgradeapi.Plan{}, genOpts.WithDisallowNil()),
		genType(&upgradetypes.MsgSoftwareUpgrade{}, &upgradeapi.MsgSoftwareUpgrade{}, genOpts.WithDisallowNil()),
		genType(&upgradetypes.MsgCancelUpgrade{}, &upgradeapi.MsgCancelUpgrade{}, genOpts),

		// vesting
		genType(&vestingtypes.BaseVestingAccount{}, &vestingapi.BaseVestingAccount{}, genOpts.WithDisallowNil()),
		genType(&vestingtypes.ContinuousVestingAccount{}, &vestingapi.ContinuousVestingAccount{}, genOpts.WithDisallowNil()),
		genType(&vestingtypes.DelayedVestingAccount{}, &vestingapi.DelayedVestingAccount{}, genOpts.WithDisallowNil()),
		genType(&vestingtypes.PeriodicVestingAccount{}, &vestingapi.PeriodicVestingAccount{}, genOpts.WithDisallowNil()),
		genType(&vestingtypes.PermanentLockedAccount{}, &vestingapi.PermanentLockedAccount{}, genOpts.WithDisallowNil()),
		genType(&vestingtypes.MsgCreateVestingAccount{}, &vestingapi.MsgCreateVestingAccount{}, genOpts),
		genType(&vestingtypes.MsgCreatePermanentLockedAccount{}, &vestingapi.MsgCreatePermanentLockedAccount{}, genOpts),
		genType(&vestingtypes.MsgCreatePeriodicVestingAccount{}, &vestingapi.MsgCreatePeriodicVestingAccount{}, genOpts),
	}
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
		auth.AppModuleBasic{}, authzmodule.AppModuleBasic{}, bank.AppModuleBasic{}, consensus.AppModuleBasic{},
		distribution.AppModuleBasic{}, evidence.AppModuleBasic{}, feegrantmodule.AppModuleBasic{},
		gov.AppModuleBasic{}, groupmodule.AppModuleBasic{}, mint.AppModuleBasic{}, params.AppModuleBasic{},
		slashing.AppModuleBasic{}, staking.AppModuleBasic{}, upgrade.AppModuleBasic{}, vesting.AppModuleBasic{})
	aj := aminojson.NewAminoJSON()

	for _, tt := range genTypes {
		name := string(tt.pulsar.ProtoReflect().Descriptor().FullName())
		t.Run(name, func(t *testing.T) {
			gen := rapidproto.MessageGenerator(tt.pulsar, tt.opts)
			fmt.Printf("testing %s\n", tt.pulsar.ProtoReflect().Descriptor().FullName())
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

				gogo := tt.gogo
				sanity := tt.pulsar

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
				gogoMsg, ok := gogo.(types.Msg)
				if !ok {
					// not signable
					return
				}

				handlerOptions := signing_testutil.HandlerArgumentOptions{
					ChainId:       "test-chain",
					Memo:          "sometestmemo",
					Msg:           tt.pulsar,
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
				require.NoError(t, txBuilder.SetMsgs([]types.Msg{gogoMsg}...))
				txBuilder.SetMemo(handlerOptions.Memo)
				txBuilder.SetFeeAmount(types.Coins{types.NewInt64Coin("uatom", 1000)})
				theTx := txBuilder.GetTx()

				legacySigningData := signing.SignerData{
					ChainID:       handlerOptions.ChainId,
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
	encCfg := testutil.MakeTestEncodingConfig(auth.AppModuleBasic{}, authzmodule.AppModuleBasic{},
		bank.AppModuleBasic{}, distribution.AppModuleBasic{}, slashing.AppModuleBasic{}, staking.AppModuleBasic{},
		vesting.AppModuleBasic{})

	aj := aminojson.NewAminoJSON()
	addr1 := types.AccAddress("addr1")
	now := time.Now()

	genericAuth, _ := codectypes.NewAnyWithValue(&authztypes.GenericAuthorization{Msg: "foo"})
	genericAuthPulsar := newAny(t, &authzapi.GenericAuthorization{Msg: "foo"})
	pubkeyAny, _ := codectypes.NewAnyWithValue(&secp256k1types.PubKey{Key: []byte("foo")})
	pubkeyAnyPulsar := newAny(t, &secp256k1.PubKey{Key: []byte("foo")})
	dec10bz, _ := types.NewDec(10).Marshal()
	int123bz, _ := types.NewInt(123).Marshal()

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
				MinSignedPerWindow:   types.NewDec(10),
			},
			pulsar: &slashingapi.Params{
				DowntimeJailDuration: &durationpb.Duration{Seconds: 1, Nanos: 7},
				MinSignedPerWindow:   dec10bz,
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
			gogo:   &gogo_testpb.IntAsString{IntAsString: types.NewInt(123)},
			pulsar: &pulsar_testpb.IntAsString{IntAsString: "123"},
		},
		"math/int_as_string/empty": {
			gogo:   &gogo_testpb.IntAsString{},
			pulsar: &pulsar_testpb.IntAsString{},
		},
		"math/int_as_bytes": {
			gogo:   &gogo_testpb.IntAsBytes{IntAsBytes: types.NewInt(123)},
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
			msg, ok := tc.gogo.(types.Msg)
			if !ok {
				// not signable
				return
			}

			handlerOptions := signing_testutil.HandlerArgumentOptions{
				ChainId:       "test-chain",
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
				ChainID:       handlerOptions.ChainId,
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
	encCfg := testutil.MakeTestEncodingConfig(staking.AppModuleBasic{})
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
