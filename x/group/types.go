package group

import (
	"fmt"
	"time"

	proto "github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/group/errors"
	"github.com/cosmos/cosmos-sdk/x/group/internal/math"
	"github.com/cosmos/cosmos-sdk/x/group/internal/orm"
)

// MaxMetadataLength defines the max length of the metadata bytes field
// for various entities within the group module
// TODO: This could be used as params once x/params is upgraded to use protobuf
const MaxMetadataLength = 255

type DecisionPolicyResult struct {
	Allow bool
	Final bool
}

// DecisionPolicy is the persistent set of rules to determine the result of election on a proposal.
type DecisionPolicy interface {
	codec.ProtoMarshaler

	ValidateBasic() error
	GetTimeout() time.Duration
	Allow(tallyResult TallyResult, totalPower string, votingDuration time.Duration) (DecisionPolicyResult, error)
	Validate(g GroupInfo) error
}

// Implements DecisionPolicy Interface
var _ DecisionPolicy = &ThresholdDecisionPolicy{}

// NewThresholdDecisionPolicy creates a threshold DecisionPolicy
func NewThresholdDecisionPolicy(threshold string, timeout time.Duration) DecisionPolicy {
	return &ThresholdDecisionPolicy{threshold, timeout}
}

func (p ThresholdDecisionPolicy) ValidateBasic() error {
	if _, err := math.NewPositiveDecFromString(p.Threshold); err != nil {
		return sdkerrors.Wrap(err, "threshold")
	}

	timeout := p.Timeout

	if timeout <= time.Nanosecond {
		return sdkerrors.Wrap(errors.ErrInvalid, "timeout")
	}
	return nil
}

// Allow allows a proposal to pass when the tallyResult of yes votes equals or exceeds the threshold before the timeout.
func (p ThresholdDecisionPolicy) Allow(tallyResult TallyResult, totalPower string, votingDuration time.Duration) (DecisionPolicyResult, error) {
	pTimeout := types.DurationProto(p.Timeout)
	timeout, err := types.DurationFromProto(pTimeout)
	if err != nil {
		return DecisionPolicyResult{}, err
	}
	if timeout <= votingDuration {
		return DecisionPolicyResult{Allow: false, Final: true}, nil
	}

	threshold, err := math.NewPositiveDecFromString(p.Threshold)
	if err != nil {
		return DecisionPolicyResult{}, err
	}
	yesCount, err := math.NewNonNegativeDecFromString(tallyResult.Yes)
	if err != nil {
		return DecisionPolicyResult{}, err
	}
	if yesCount.Cmp(threshold) >= 0 {
		return DecisionPolicyResult{Allow: true, Final: true}, nil
	}

	totalPowerDec, err := math.NewNonNegativeDecFromString(totalPower)
	if err != nil {
		return DecisionPolicyResult{}, err
	}
	totalCounts, err := tallyResult.TotalCounts()
	if err != nil {
		return DecisionPolicyResult{}, err
	}
	undecided, err := math.SubNonNegative(totalPowerDec, totalCounts)
	if err != nil {
		return DecisionPolicyResult{}, err
	}
	sum, err := yesCount.Add(undecided)
	if err != nil {
		return DecisionPolicyResult{}, err
	}
	if sum.Cmp(threshold) < 0 {
		return DecisionPolicyResult{Allow: false, Final: true}, nil
	}
	return DecisionPolicyResult{Allow: false, Final: false}, nil
}

// Validate returns an error if policy threshold is greater than the total group weight
func (p *ThresholdDecisionPolicy) Validate(g GroupInfo) error {
	threshold, err := math.NewPositiveDecFromString(p.Threshold)
	if err != nil {
		return sdkerrors.Wrap(err, "threshold")
	}
	totalWeight, err := math.NewNonNegativeDecFromString(g.TotalWeight)
	if err != nil {
		return sdkerrors.Wrap(err, "group total weight")
	}
	if threshold.Cmp(totalWeight) > 0 {
		return sdkerrors.Wrapf(errors.ErrInvalid, "policy threshold %s should not be greater than the total group weight %s", p.Threshold, g.TotalWeight)
	}
	return nil
}

var _ orm.Validateable = GroupPolicyInfo{}

// NewGroupPolicyInfo creates a new GroupPolicyInfo instance
func NewGroupPolicyInfo(address sdk.AccAddress, group uint64, admin sdk.AccAddress, metadata []byte,
	version uint64, decisionPolicy DecisionPolicy, createdAt time.Time) (GroupPolicyInfo, error) {
	p := GroupPolicyInfo{
		Address:   address.String(),
		GroupId:   group,
		Admin:     admin.String(),
		Metadata:  metadata,
		Version:   version,
		CreatedAt: createdAt,
	}

	err := p.SetDecisionPolicy(decisionPolicy)
	if err != nil {
		return GroupPolicyInfo{}, err
	}

	return p, nil
}

func (g *GroupPolicyInfo) SetDecisionPolicy(decisionPolicy DecisionPolicy) error {
	msg, ok := decisionPolicy.(proto.Message)
	if !ok {
		return fmt.Errorf("can't proto marshal %T", msg)
	}
	any, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return err
	}
	g.DecisionPolicy = any
	return nil
}

func (g GroupPolicyInfo) GetDecisionPolicy() DecisionPolicy {
	decisionPolicy, ok := g.DecisionPolicy.GetCachedValue().(DecisionPolicy)
	if !ok {
		return nil
	}
	return decisionPolicy
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (g GroupPolicyInfo) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var decisionPolicy DecisionPolicy
	return unpacker.UnpackAny(g.DecisionPolicy, &decisionPolicy)
}

func (g GroupInfo) PrimaryKeyFields() []interface{} {
	return []interface{}{g.GroupId}
}

func (g GroupInfo) ValidateBasic() error {
	if g.GroupId == 0 {
		return sdkerrors.Wrap(errors.ErrEmpty, "group's GroupId")
	}

	_, err := sdk.AccAddressFromBech32(g.Admin)
	if err != nil {
		return sdkerrors.Wrap(err, "admin")
	}

	if _, err := math.NewNonNegativeDecFromString(g.TotalWeight); err != nil {
		return sdkerrors.Wrap(err, "total weight")
	}
	if g.Version == 0 {
		return sdkerrors.Wrap(errors.ErrEmpty, "version")
	}
	return nil
}

func (g GroupPolicyInfo) PrimaryKeyFields() []interface{} {
	addr, err := sdk.AccAddressFromBech32(g.Address)
	if err != nil {
		panic(err)
	}
	return []interface{}{addr.Bytes()}
}

func (g Proposal) PrimaryKeyFields() []interface{} {
	return []interface{}{g.ProposalId}
}

func (g GroupPolicyInfo) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(g.Admin)
	if err != nil {
		return sdkerrors.Wrap(err, "group policy admin")
	}
	_, err = sdk.AccAddressFromBech32(g.Address)
	if err != nil {
		return sdkerrors.Wrap(err, "group policy account address")
	}

	if g.GroupId == 0 {
		return sdkerrors.Wrap(errors.ErrEmpty, "group policy's group id")
	}
	if g.Version == 0 {
		return sdkerrors.Wrap(errors.ErrEmpty, "group policy version")
	}
	policy := g.GetDecisionPolicy()

	if policy == nil {
		return sdkerrors.Wrap(errors.ErrEmpty, "group policy's decision policy")
	}
	if err := policy.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "group policy's decision policy")
	}
	return nil
}

func (g GroupMember) PrimaryKeyFields() []interface{} {
	addr, err := sdk.AccAddressFromBech32(g.Member.Address)
	if err != nil {
		panic(err)
	}
	return []interface{}{g.GroupId, addr.Bytes()}
}

func (g GroupMember) ValidateBasic() error {
	if g.GroupId == 0 {
		return sdkerrors.Wrap(errors.ErrEmpty, "group member's group id")
	}

	err := g.Member.ValidateBasic()
	if err != nil {
		return sdkerrors.Wrap(err, "group member")
	}
	return nil
}

func (p Proposal) ValidateBasic() error {

	if p.ProposalId == 0 {
		return sdkerrors.Wrap(errors.ErrEmpty, "proposal id")
	}
	_, err := sdk.AccAddressFromBech32(p.Address)
	if err != nil {
		return sdkerrors.Wrap(err, "proposal group policy address")
	}
	if p.GroupVersion == 0 {
		return sdkerrors.Wrap(errors.ErrEmpty, "proposal group version")
	}
	if p.GroupPolicyVersion == 0 {
		return sdkerrors.Wrap(errors.ErrEmpty, "proposal group policy version")
	}
	_, err = p.FinalTallyResult.GetYesCount()
	if err != nil {
		return sdkerrors.Wrap(err, "proposal FinalTallyResult yes count")
	}
	_, err = p.FinalTallyResult.GetNoCount()
	if err != nil {
		return sdkerrors.Wrap(err, "proposal FinalTallyResult no count")
	}
	_, err = p.FinalTallyResult.GetAbstainCount()
	if err != nil {
		return sdkerrors.Wrap(err, "proposal FinalTallyResult abstain count")
	}
	_, err = p.FinalTallyResult.GetNoWithVetoCount()
	if err != nil {
		return sdkerrors.Wrap(err, "proposal FinalTallyResult veto count")
	}
	return nil
}

func (v Vote) PrimaryKeyFields() []interface{} {
	addr, err := sdk.AccAddressFromBech32(v.Voter)
	if err != nil {
		panic(err)
	}
	return []interface{}{v.ProposalId, addr.Bytes()}
}

var _ orm.Validateable = Vote{}

func (v Vote) ValidateBasic() error {

	_, err := sdk.AccAddressFromBech32(v.Voter)
	if err != nil {
		return sdkerrors.Wrap(err, "voter")
	}
	if v.ProposalId == 0 {
		return sdkerrors.Wrap(errors.ErrEmpty, "voter ProposalId")
	}
	if v.Option == VOTE_OPTION_UNSPECIFIED {
		return sdkerrors.Wrap(errors.ErrEmpty, "voter vote option")
	}
	if _, ok := VoteOption_name[int32(v.Option)]; !ok {
		return sdkerrors.Wrap(errors.ErrInvalid, "vote option")
	}
	return nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (q QueryGroupPoliciesByGroupResponse) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return unpackGroupPolicies(unpacker, q.GroupPolicies)
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (q QueryGroupPoliciesByAdminResponse) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return unpackGroupPolicies(unpacker, q.GroupPolicies)
}

func unpackGroupPolicies(unpacker codectypes.AnyUnpacker, accs []*GroupPolicyInfo) error {
	for _, g := range accs {
		err := g.UnpackInterfaces(unpacker)
		if err != nil {
			return err
		}
	}

	return nil
}

type operation func(x, y math.Dec) (math.Dec, error)

func (t *TallyResult) operation(vote Vote, weight string, op operation) error {
	weightDec, err := math.NewPositiveDecFromString(weight)
	if err != nil {
		return err
	}

	yesCount, err := t.GetYesCount()
	if err != nil {
		return sdkerrors.Wrap(err, "yes count")
	}
	noCount, err := t.GetNoCount()
	if err != nil {
		return sdkerrors.Wrap(err, "no count")
	}
	abstainCount, err := t.GetAbstainCount()
	if err != nil {
		return sdkerrors.Wrap(err, "abstain count")
	}
	vetoCount, err := t.GetNoWithVetoCount()
	if err != nil {
		return sdkerrors.Wrap(err, "veto count")
	}

	switch vote.Option {
	case VOTE_OPTION_YES:
		yesCount, err := op(yesCount, weightDec)
		if err != nil {
			return sdkerrors.Wrap(err, "yes count")
		}
		t.Yes = yesCount.String()
	case VOTE_OPTION_NO:
		noCount, err := op(noCount, weightDec)
		if err != nil {
			return sdkerrors.Wrap(err, "no count")
		}
		t.No = noCount.String()
	case VOTE_OPTION_ABSTAIN:
		abstainCount, err := op(abstainCount, weightDec)
		if err != nil {
			return sdkerrors.Wrap(err, "abstain count")
		}
		t.Abstain = abstainCount.String()
	case VOTE_OPTION_NO_WITH_VETO:
		vetoCount, err := op(vetoCount, weightDec)
		if err != nil {
			return sdkerrors.Wrap(err, "veto count")
		}
		t.NoWithVeto = vetoCount.String()
	default:
		return sdkerrors.Wrapf(errors.ErrInvalid, "unknown vote option %s", vote.Option.String())
	}
	return nil
}

func (t TallyResult) GetYesCount() (math.Dec, error) {
	yesCount, err := math.NewNonNegativeDecFromString(t.Yes)
	if err != nil {
		return math.Dec{}, err
	}
	return yesCount, nil
}

func (t TallyResult) GetNoCount() (math.Dec, error) {
	noCount, err := math.NewNonNegativeDecFromString(t.No)
	if err != nil {
		return math.Dec{}, err
	}
	return noCount, nil
}

func (t TallyResult) GetAbstainCount() (math.Dec, error) {
	abstainCount, err := math.NewNonNegativeDecFromString(t.Abstain)
	if err != nil {
		return math.Dec{}, err
	}
	return abstainCount, nil
}

func (t TallyResult) GetNoWithVetoCount() (math.Dec, error) {
	vetoCount, err := math.NewNonNegativeDecFromString(t.NoWithVeto)
	if err != nil {
		return math.Dec{}, err
	}
	return vetoCount, nil
}

func (t *TallyResult) Add(vote Vote, weight string) error {
	if err := t.operation(vote, weight, math.Add); err != nil {
		return err
	}
	return nil
}

// TotalCounts is the sum of all weights.
func (t TallyResult) TotalCounts() (math.Dec, error) {
	yesCount, err := t.GetYesCount()
	if err != nil {
		return math.Dec{}, sdkerrors.Wrap(err, "yes count")
	}
	noCount, err := t.GetNoCount()
	if err != nil {
		return math.Dec{}, sdkerrors.Wrap(err, "no count")
	}
	abstainCount, err := t.GetAbstainCount()
	if err != nil {
		return math.Dec{}, sdkerrors.Wrap(err, "abstain count")
	}
	vetoCount, err := t.GetNoWithVetoCount()
	if err != nil {
		return math.Dec{}, sdkerrors.Wrap(err, "veto count")
	}

	totalCounts := math.NewDecFromInt64(0)
	totalCounts, err = totalCounts.Add(yesCount)
	if err != nil {
		return math.Dec{}, err
	}
	totalCounts, err = totalCounts.Add(noCount)
	if err != nil {
		return math.Dec{}, err
	}
	totalCounts, err = totalCounts.Add(abstainCount)
	if err != nil {
		return math.Dec{}, err
	}
	totalCounts, err = totalCounts.Add(vetoCount)
	if err != nil {
		return math.Dec{}, err
	}
	return totalCounts, nil
}

// VoteOptionFromString returns a VoteOption from a string. It returns an error
// if the string is invalid.
func VoteOptionFromString(str string) (VoteOption, error) {
	vo, ok := VoteOption_value[str]
	if !ok {
		return VOTE_OPTION_UNSPECIFIED, fmt.Errorf("'%s' is not a valid vote option", str)
	}
	return VoteOption(vo), nil
}
