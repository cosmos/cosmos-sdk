package rapidgen

import (
	"fmt"

	cosmos_proto "github.com/cosmos/cosmos-proto"
	"github.com/cosmos/cosmos-proto/rapidproto"
	gogoproto "github.com/cosmos/gogoproto/proto"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"pgregory.net/rapid"

	authapi "cosmossdk.io/api/cosmos/auth/v1beta1"
	authzapi "cosmossdk.io/api/cosmos/authz/v1beta1"
	bankapi "cosmossdk.io/api/cosmos/bank/v1beta1"
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
	upgradeapi "cosmossdk.io/api/cosmos/upgrade/v1beta1"
	vestingapi "cosmossdk.io/api/cosmos/vesting/v1beta1"

	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	authztypes "github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	feegranttypes "github.com/cosmos/cosmos-sdk/x/feegrant"
	gov_v1_types "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	gov_v1beta1_types "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	grouptypes "github.com/cosmos/cosmos-sdk/x/group" //nolint:staticcheck // deprecated and to be removed
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

type GeneratedType struct {
	Pulsar proto.Message
	Gogo   gogoproto.Message
	Opts   rapidproto.GeneratorOptions
}

func GenType(gogo gogoproto.Message, pulsar proto.Message, opts rapidproto.GeneratorOptions) GeneratedType {
	return GeneratedType{
		Pulsar: pulsar,
		Gogo:   gogo,
		Opts:   opts,
	}
}

func WithDecisionPolicy(opts rapidproto.GeneratorOptions) rapidproto.GeneratorOptions {
	return opts.
		WithAnyTypes(
			&groupapi.ThresholdDecisionPolicy{},
			&groupapi.PercentageDecisionPolicy{}).
		WithDisallowNil().
		WithInterfaceHint("cosmos.group.v1.DecisionPolicy", &groupapi.ThresholdDecisionPolicy{}).
		WithInterfaceHint("cosmos.group.v1.DecisionPolicy", &groupapi.PercentageDecisionPolicy{})
}

func GeneratorFieldMapper(t *rapid.T, field protoreflect.FieldDescriptor, name string) (protoreflect.Value, bool) {
	opts := field.Options()
	if proto.HasExtension(opts, cosmos_proto.E_Scalar) {
		scalar := proto.GetExtension(opts, cosmos_proto.E_Scalar).(string)
		switch scalar {
		case "cosmos.Int":
			i32 := rapid.Int32().Draw(t, name)
			return protoreflect.ValueOfString(fmt.Sprintf("%d", i32)), true
		case "cosmos.Dec":
			if field.Kind() == protoreflect.BytesKind {
				return protoreflect.ValueOfBytes([]byte{}), true
			}

			return protoreflect.ValueOfString(""), true
		}
	}

	return protoreflect.Value{}, false
}

var (
	GenOpts = rapidproto.GeneratorOptions{
		Resolver:  protoregistry.GlobalTypes,
		FieldMaps: []rapidproto.FieldMapper{GeneratorFieldMapper},
	}
	SignableTypes = []GeneratedType{
		// auth
		GenType(&authtypes.MsgUpdateParams{}, &authapi.MsgUpdateParams{}, GenOpts.WithDisallowNil()),

		// authz
		GenType(&authztypes.MsgGrant{}, &authzapi.MsgGrant{},
			GenOpts.WithAnyTypes(&authzapi.GenericAuthorization{}).
				WithInterfaceHint("cosmos.authz.v1beta1.Authorization", &authzapi.GenericAuthorization{}).
				WithDisallowNil(),
		),
		GenType(&authztypes.MsgExec{}, &authzapi.MsgExec{},
			GenOpts.WithAnyTypes(&authzapi.MsgGrant{}, &authzapi.GenericAuthorization{}).
				WithDisallowNil().
				WithInterfaceHint("cosmos.authz.v1beta1.Authorization", &authzapi.GenericAuthorization{}).
				WithInterfaceHint("cosmos.base.v1beta1.Msg", &authzapi.MsgGrant{}),
		),

		// bank
		GenType(&banktypes.MsgSend{}, &bankapi.MsgSend{}, GenOpts.WithDisallowNil()),
		GenType(&banktypes.MsgMultiSend{}, &bankapi.MsgMultiSend{}, GenOpts.WithDisallowNil()),
		GenType(&banktypes.MsgUpdateParams{}, &bankapi.MsgUpdateParams{}, GenOpts.WithDisallowNil()),
		GenType(&banktypes.MsgSetSendEnabled{}, &bankapi.MsgSetSendEnabled{}, GenOpts),

		// consensus
		GenType(&consensustypes.MsgUpdateParams{}, &consensusapi.MsgUpdateParams{}, GenOpts.WithDisallowNil()),

		// distribution
		GenType(&disttypes.MsgWithdrawDelegatorReward{}, &distapi.MsgWithdrawDelegatorReward{}, GenOpts),
		GenType(&disttypes.MsgWithdrawValidatorCommission{}, &distapi.MsgWithdrawValidatorCommission{}, GenOpts),
		GenType(&disttypes.MsgSetWithdrawAddress{}, &distapi.MsgSetWithdrawAddress{}, GenOpts),
		GenType(&disttypes.MsgFundCommunityPool{}, &distapi.MsgFundCommunityPool{}, GenOpts),
		GenType(&disttypes.MsgUpdateParams{}, &distapi.MsgUpdateParams{}, GenOpts.WithDisallowNil()),
		GenType(&disttypes.MsgCommunityPoolSpend{}, &distapi.MsgCommunityPoolSpend{}, GenOpts),
		GenType(&disttypes.MsgDepositValidatorRewardsPool{}, &distapi.MsgDepositValidatorRewardsPool{}, GenOpts),

		// evidence
		GenType(&evidencetypes.MsgSubmitEvidence{}, &evidenceapi.MsgSubmitEvidence{},
			GenOpts.WithAnyTypes(&evidenceapi.Equivocation{}).
				WithDisallowNil().
				WithInterfaceHint("cosmos.evidence.v1beta1.Evidence", &evidenceapi.Equivocation{})),

		// feegrant
		GenType(&feegranttypes.MsgGrantAllowance{}, &feegrantapi.MsgGrantAllowance{},
			GenOpts.WithDisallowNil().
				WithAnyTypes(
					&feegrantapi.BasicAllowance{},
					&feegrantapi.PeriodicAllowance{}).
				WithInterfaceHint("cosmos.feegrant.v1beta1.FeeAllowanceI", &feegrantapi.BasicAllowance{}).
				WithInterfaceHint("cosmos.feegrant.v1beta1.FeeAllowanceI", &feegrantapi.PeriodicAllowance{}),
		),
		GenType(&feegranttypes.MsgRevokeAllowance{}, &feegrantapi.MsgRevokeAllowance{}, GenOpts),

		// gov v1beta1
		GenType(&gov_v1beta1_types.MsgSubmitProposal{}, &gov_v1beta1_api.MsgSubmitProposal{},
			GenOpts.WithAnyTypes(&gov_v1beta1_api.TextProposal{}).
				WithDisallowNil().
				WithInterfaceHint("cosmos.gov.v1beta1.Content", &gov_v1beta1_api.TextProposal{}),
		),
		GenType(&gov_v1beta1_types.MsgDeposit{}, &gov_v1beta1_api.MsgDeposit{}, GenOpts),
		GenType(&gov_v1beta1_types.MsgVote{}, &gov_v1beta1_api.MsgVote{}, GenOpts),
		GenType(&gov_v1beta1_types.MsgVoteWeighted{}, &gov_v1beta1_api.MsgVoteWeighted{}, GenOpts),

		// gov v1
		GenType(&gov_v1_types.MsgSubmitProposal{}, &gov_v1_api.MsgSubmitProposal{},
			GenOpts.WithAnyTypes(&gov_v1_api.MsgVote{}, &gov_v1_api.MsgVoteWeighted{}, &gov_v1_api.MsgDeposit{},
				&gov_v1_api.MsgExecLegacyContent{}, &gov_v1_api.MsgUpdateParams{}).
				WithInterfaceHint("cosmos.gov.v1beta1.Content", &gov_v1beta1_api.TextProposal{}).
				WithDisallowNil(),
		),
		GenType(&gov_v1_types.MsgDeposit{}, &gov_v1_api.MsgDeposit{}, GenOpts),
		GenType(&gov_v1_types.MsgVote{}, &gov_v1_api.MsgVote{}, GenOpts),
		GenType(&gov_v1_types.MsgVoteWeighted{}, &gov_v1_api.MsgVoteWeighted{}, GenOpts),
		GenType(&gov_v1_types.MsgExecLegacyContent{}, &gov_v1_api.MsgExecLegacyContent{},
			GenOpts.WithAnyTypes(&gov_v1beta1_api.TextProposal{}).
				WithDisallowNil().
				WithInterfaceHint("cosmos.gov.v1beta1.Content", &gov_v1beta1_api.TextProposal{})),
		GenType(&gov_v1_types.MsgUpdateParams{}, &gov_v1_api.MsgUpdateParams{}, GenOpts.WithDisallowNil()),

		// group
		GenType(&grouptypes.MsgCreateGroup{}, &groupapi.MsgCreateGroup{}, GenOpts),
		GenType(&grouptypes.MsgUpdateGroupMembers{}, &groupapi.MsgUpdateGroupMembers{}, GenOpts),
		GenType(&grouptypes.MsgUpdateGroupAdmin{}, &groupapi.MsgUpdateGroupAdmin{}, GenOpts),
		GenType(&grouptypes.MsgUpdateGroupMetadata{}, &groupapi.MsgUpdateGroupMetadata{}, GenOpts),
		GenType(&grouptypes.MsgCreateGroupWithPolicy{}, &groupapi.MsgCreateGroupWithPolicy{},
			WithDecisionPolicy(GenOpts)),
		GenType(&grouptypes.MsgCreateGroupPolicy{}, &groupapi.MsgCreateGroupPolicy{},
			WithDecisionPolicy(GenOpts)),
		GenType(&grouptypes.MsgUpdateGroupPolicyAdmin{}, &groupapi.MsgUpdateGroupPolicyAdmin{}, GenOpts),
		GenType(&grouptypes.MsgUpdateGroupPolicyDecisionPolicy{}, &groupapi.MsgUpdateGroupPolicyDecisionPolicy{},
			WithDecisionPolicy(GenOpts)),
		GenType(&grouptypes.MsgUpdateGroupPolicyMetadata{}, &groupapi.MsgUpdateGroupPolicyMetadata{}, GenOpts),
		GenType(&grouptypes.MsgSubmitProposal{}, &groupapi.MsgSubmitProposal{},
			GenOpts.WithDisallowNil().
				WithAnyTypes(&groupapi.MsgCreateGroup{}, &groupapi.MsgUpdateGroupMembers{}).
				WithInterfaceHint("cosmos.base.v1beta1.Msg", &groupapi.MsgCreateGroup{}).
				WithInterfaceHint("cosmos.base.v1beta1.Msg", &groupapi.MsgUpdateGroupMembers{}),
		),
		GenType(&grouptypes.MsgVote{}, &groupapi.MsgVote{}, GenOpts),
		GenType(&grouptypes.MsgExec{}, &groupapi.MsgExec{}, GenOpts),
		GenType(&grouptypes.MsgLeaveGroup{}, &groupapi.MsgLeaveGroup{}, GenOpts),

		// mint
		GenType(&minttypes.MsgUpdateParams{}, &mintapi.MsgUpdateParams{}, GenOpts.WithDisallowNil()),

		// slashing
		GenType(&slashingtypes.MsgUnjail{}, &slashingapi.MsgUnjail{}, GenOpts),
		GenType(&slashingtypes.MsgUpdateParams{}, &slashingapi.MsgUpdateParams{}, GenOpts.WithDisallowNil()),

		// staking
		GenType(&stakingtypes.MsgCreateValidator{}, &stakingapi.MsgCreateValidator{},
			GenOpts.WithDisallowNil().
				WithAnyTypes(&ed25519.PubKey{}).
				WithInterfaceHint("cosmos.crypto.PubKey", &ed25519.PubKey{}),
		),
		GenType(&stakingtypes.MsgEditValidator{}, &stakingapi.MsgEditValidator{}, GenOpts.WithDisallowNil()),
		GenType(&stakingtypes.MsgDelegate{}, &stakingapi.MsgDelegate{}, GenOpts.WithDisallowNil()),
		GenType(&stakingtypes.MsgUndelegate{}, &stakingapi.MsgUndelegate{}, GenOpts.WithDisallowNil()),
		GenType(&stakingtypes.MsgBeginRedelegate{}, &stakingapi.MsgBeginRedelegate{}, GenOpts.WithDisallowNil()),
		GenType(&stakingtypes.MsgUpdateParams{}, &stakingapi.MsgUpdateParams{}, GenOpts.WithDisallowNil()),

		// upgrade
		GenType(&upgradetypes.MsgSoftwareUpgrade{}, &upgradeapi.MsgSoftwareUpgrade{}, GenOpts.WithDisallowNil()),
		GenType(&upgradetypes.MsgCancelUpgrade{}, &upgradeapi.MsgCancelUpgrade{}, GenOpts),

		// vesting
		GenType(&vestingtypes.MsgCreateVestingAccount{}, &vestingapi.MsgCreateVestingAccount{}, GenOpts),
		GenType(&vestingtypes.MsgCreatePermanentLockedAccount{}, &vestingapi.MsgCreatePermanentLockedAccount{}, GenOpts),
		GenType(&vestingtypes.MsgCreatePeriodicVestingAccount{}, &vestingapi.MsgCreatePeriodicVestingAccount{}, GenOpts),
	}
	NonsignableTypes = []GeneratedType{
		GenType(&authtypes.Params{}, &authapi.Params{}, GenOpts),
		GenType(&authtypes.BaseAccount{}, &authapi.BaseAccount{}, GenOpts.WithAnyTypes(&ed25519.PubKey{})),
		GenType(&authtypes.ModuleCredential{}, &authapi.ModuleCredential{}, GenOpts),

		GenType(&authztypes.GenericAuthorization{}, &authzapi.GenericAuthorization{}, GenOpts),
		GenType(&authztypes.Grant{}, &authzapi.Grant{},
			GenOpts.WithAnyTypes(&authzapi.GenericAuthorization{}).
				WithDisallowNil().
				WithInterfaceHint("cosmos.authz.v1beta1.Authorization", &authzapi.GenericAuthorization{}),
		),

		GenType(&banktypes.SendAuthorization{}, &bankapi.SendAuthorization{}, GenOpts),
		GenType(&banktypes.Params{}, &bankapi.Params{}, GenOpts),

		// crypto
		GenType(&multisig.LegacyAminoPubKey{}, &multisigapi.LegacyAminoPubKey{},
			GenOpts.WithAnyTypes(&ed25519.PubKey{}, &secp256k1.PubKey{})),

		GenType(&disttypes.Params{}, &distapi.Params{}, GenOpts),

		GenType(&evidencetypes.Equivocation{}, &evidenceapi.Equivocation{}, GenOpts.WithDisallowNil()),

		GenType(&feegranttypes.BasicAllowance{}, &feegrantapi.BasicAllowance{}, GenOpts.WithDisallowNil()),
		GenType(&feegranttypes.PeriodicAllowance{}, &feegrantapi.PeriodicAllowance{}, GenOpts.WithDisallowNil()),
		GenType(&feegranttypes.AllowedMsgAllowance{}, &feegrantapi.AllowedMsgAllowance{},
			GenOpts.WithDisallowNil().
				WithAnyTypes(
					&feegrantapi.BasicAllowance{},
					&feegrantapi.PeriodicAllowance{}).
				WithInterfaceHint("cosmos.feegrant.v1beta1.FeeAllowanceI", &feegrantapi.BasicAllowance{}).
				WithInterfaceHint("cosmos.feegrant.v1beta1.FeeAllowanceI", &feegrantapi.PeriodicAllowance{}),
		),

		GenType(&gov_v1beta1_types.TextProposal{}, &gov_v1beta1_api.TextProposal{}, GenOpts),

		GenType(&minttypes.Params{}, &mintapi.Params{}, GenOpts),

		// params
		GenType(&proposal.ParameterChangeProposal{}, &paramsapi.ParameterChangeProposal{}, GenOpts),

		GenType(&slashingtypes.Params{}, &slashingapi.Params{}, GenOpts.WithDisallowNil()),

		GenType(&stakingtypes.StakeAuthorization{}, &stakingapi.StakeAuthorization{}, GenOpts),

		GenType(&upgradetypes.CancelSoftwareUpgradeProposal{}, &upgradeapi.CancelSoftwareUpgradeProposal{}, GenOpts),       //nolint:staticcheck // testing registration of legacy deprecated type
		GenType(&upgradetypes.SoftwareUpgradeProposal{}, &upgradeapi.SoftwareUpgradeProposal{}, GenOpts.WithDisallowNil()), //nolint:staticcheck // testing registration of legacy deprecated type
		GenType(&upgradetypes.Plan{}, &upgradeapi.Plan{}, GenOpts.WithDisallowNil()),

		GenType(&vestingtypes.BaseVestingAccount{}, &vestingapi.BaseVestingAccount{}, GenOpts.WithDisallowNil()),
		GenType(&vestingtypes.ContinuousVestingAccount{}, &vestingapi.ContinuousVestingAccount{}, GenOpts.WithDisallowNil()),
		GenType(&vestingtypes.DelayedVestingAccount{}, &vestingapi.DelayedVestingAccount{}, GenOpts.WithDisallowNil()),
		GenType(&vestingtypes.PermanentLockedAccount{}, &vestingapi.PermanentLockedAccount{}, GenOpts.WithDisallowNil()),
		GenType(&vestingtypes.PeriodicVestingAccount{}, &vestingapi.PeriodicVestingAccount{}, GenOpts.WithDisallowNil()),
	}
	DefaultGeneratedTypes = append(SignableTypes, NonsignableTypes...)
)
