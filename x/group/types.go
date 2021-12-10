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
	Allow(tally Tally, totalPower string, votingDuration time.Duration) (DecisionPolicyResult, error)
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

// Allow allows a proposal to pass when the tally of yes votes equals or exceeds the threshold before the timeout.
func (p ThresholdDecisionPolicy) Allow(tally Tally, totalPower string, votingDuration time.Duration) (DecisionPolicyResult, error) {
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
	yesCount, err := math.NewNonNegativeDecFromString(tally.YesCount)
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
	totalCounts, err := tally.TotalCounts()
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
		return sdkerrors.Wrap(errors.ErrInvalid, "policy threshold should not be greater than the total group weight")
	}
	return nil
}

var _ orm.Validateable = GroupAccountInfo{}

// NewGroupAccountInfo creates a new GroupAccountInfo instance
func NewGroupAccountInfo(address sdk.AccAddress, group uint64, admin sdk.AccAddress, metadata []byte,
	version uint64, decisionPolicy DecisionPolicy) (GroupAccountInfo, error) {
	p := GroupAccountInfo{
		Address:  address.String(),
		GroupId:  group,
		Admin:    admin.String(),
		Metadata: metadata,
		Version:  version,
	}

	err := p.SetDecisionPolicy(decisionPolicy)
	if err != nil {
		return GroupAccountInfo{}, err
	}

	return p, nil
}

func (g *GroupAccountInfo) SetDecisionPolicy(decisionPolicy DecisionPolicy) error {
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

func (g GroupAccountInfo) GetDecisionPolicy() DecisionPolicy {
	decisionPolicy, ok := g.DecisionPolicy.GetCachedValue().(DecisionPolicy)
	if !ok {
		return nil
	}
	return decisionPolicy
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (g GroupAccountInfo) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var decisionPolicy DecisionPolicy
	return unpacker.UnpackAny(g.DecisionPolicy, &decisionPolicy)
}

func (g GroupAccountInfo) PrimaryKeyFields() []interface{} {
	addr, err := sdk.AccAddressFromBech32(g.Address)
	if err != nil {
		panic(err)
	}
	return []interface{}{addr.Bytes()}
}

func (g GroupAccountInfo) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(g.Admin)
	if err != nil {
		return sdkerrors.Wrap(err, "admin")
	}

	_, err = sdk.AccAddressFromBech32(g.Address)
	if err != nil {
		return sdkerrors.Wrap(err, "group account")
	}

	if g.GroupId == 0 {
		return sdkerrors.Wrap(errors.ErrEmpty, "group")
	}
	if g.Version == 0 {
		return sdkerrors.Wrap(errors.ErrEmpty, "version")
	}
	policy := g.GetDecisionPolicy()

	if policy == nil {
		return sdkerrors.Wrap(errors.ErrEmpty, "policy")
	}
	if err := policy.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "policy")
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
		return sdkerrors.Wrap(errors.ErrEmpty, "group")
	}

	err := g.Member.ValidateBasic()
	if err != nil {
		return sdkerrors.Wrap(err, "member")
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

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (q QueryGroupAccountsByGroupResponse) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return unpackGroupAccounts(unpacker, q.GroupAccounts)
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (q QueryGroupAccountsByAdminResponse) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return unpackGroupAccounts(unpacker, q.GroupAccounts)
}

func unpackGroupAccounts(unpacker codectypes.AnyUnpacker, accs []*GroupAccountInfo) error {
	for _, g := range accs {
		err := g.UnpackInterfaces(unpacker)
		if err != nil {
			return err
		}
	}

	return nil
}

type operation func(x, y math.Dec) (math.Dec, error)

func (t *Tally) operation(vote Vote, weight string, op operation) error {
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
	vetoCount, err := t.GetVetoCount()
	if err != nil {
		return sdkerrors.Wrap(err, "veto count")
	}

	switch vote.Choice {
	case Choice_CHOICE_YES:
		yesCount, err := op(yesCount, weightDec)
		if err != nil {
			return sdkerrors.Wrap(err, "yes count")
		}
		t.YesCount = yesCount.String()
	case Choice_CHOICE_NO:
		noCount, err := op(noCount, weightDec)
		if err != nil {
			return sdkerrors.Wrap(err, "no count")
		}
		t.NoCount = noCount.String()
	case Choice_CHOICE_ABSTAIN:
		abstainCount, err := op(abstainCount, weightDec)
		if err != nil {
			return sdkerrors.Wrap(err, "abstain count")
		}
		t.AbstainCount = abstainCount.String()
	case Choice_CHOICE_VETO:
		vetoCount, err := op(vetoCount, weightDec)
		if err != nil {
			return sdkerrors.Wrap(err, "veto count")
		}
		t.VetoCount = vetoCount.String()
	default:
		return sdkerrors.Wrapf(errors.ErrInvalid, "unknown choice %s", vote.Choice.String())
	}
	return nil
}

func (t Tally) GetYesCount() (math.Dec, error) {
	yesCount, err := math.NewNonNegativeDecFromString(t.YesCount)
	if err != nil {
		return math.Dec{}, err
	}
	return yesCount, nil
}

func (t Tally) GetNoCount() (math.Dec, error) {
	noCount, err := math.NewNonNegativeDecFromString(t.NoCount)
	if err != nil {
		return math.Dec{}, err
	}
	return noCount, nil
}

func (t Tally) GetAbstainCount() (math.Dec, error) {
	abstainCount, err := math.NewNonNegativeDecFromString(t.AbstainCount)
	if err != nil {
		return math.Dec{}, err
	}
	return abstainCount, nil
}

func (t Tally) GetVetoCount() (math.Dec, error) {
	vetoCount, err := math.NewNonNegativeDecFromString(t.VetoCount)
	if err != nil {
		return math.Dec{}, err
	}
	return vetoCount, nil
}

func (t *Tally) Add(vote Vote, weight string) error {
	if err := t.operation(vote, weight, math.Add); err != nil {
		return err
	}
	return nil
}

// TotalCounts is the sum of all weights.
func (t Tally) TotalCounts() (math.Dec, error) {
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
	vetoCount, err := t.GetVetoCount()
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
