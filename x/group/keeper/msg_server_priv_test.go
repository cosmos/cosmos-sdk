package keeper

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	coreaddress "cosmossdk.io/core/address"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/group"
	"cosmossdk.io/x/group/internal/math"

	"github.com/cosmos/cosmos-sdk/codec/address"
	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
)

func TestDoTallyAndUpdate(t *testing.T) {
	addrCodec := address.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix())
	var (
		myAddr, _      = addrCodec.BytesToString(sdk.AccAddress(bytes.Repeat([]byte{0x01}, 20)))
		myOtherAddr, _ = addrCodec.BytesToString(sdk.AccAddress(bytes.Repeat([]byte{0x02}, 20)))
	)
	encCfg := moduletestutil.MakeTestEncodingConfig(codectestutil.CodecOptions{})
	group.RegisterInterfaces(encCfg.InterfaceRegistry)

	storeKey := storetypes.NewKVStoreKey(group.StoreKey)
	testCtx := testutil.DefaultContextWithDB(t, storeKey, storetypes.NewTransientStoreKey("transient_test"))
	myAccountKeeper := &mockAccountKeeper{
		AddressCodecFn: func() coreaddress.Codec {
			return addrCodec
		},
	}
	env := runtime.NewEnvironment(runtime.NewKVStoreService(storeKey), log.NewNopLogger())

	groupKeeper := NewKeeper(env, encCfg.Codec, myAccountKeeper, group.DefaultConfig())
	noEventsFn := func(proposalID uint64) sdk.Events { return sdk.Events{} }
	type memberVote struct {
		address string
		weight  string
		option  group.VoteOption
	}
	specs := map[string]struct {
		votes           []memberVote
		policy          group.DecisionPolicy
		expStatus       group.ProposalStatus
		expVotesCleared bool
		expEvents       func(proposalID uint64) sdk.Events
	}{
		"proposal accepted": {
			votes: []memberVote{
				{address: myAddr, option: group.VOTE_OPTION_YES, weight: "2"},
				{address: myOtherAddr, option: group.VOTE_OPTION_NO, weight: "1"},
			},
			policy: mockDecisionPolicy{
				AllowFn: func(tallyResult group.TallyResult, totalPower string) (group.DecisionPolicyResult, error) {
					return group.DecisionPolicyResult{Allow: true, Final: true}, nil
				},
			},
			expStatus:       group.PROPOSAL_STATUS_ACCEPTED,
			expVotesCleared: true,
			expEvents:       noEventsFn,
		},
		"proposal rejected": {
			votes: []memberVote{
				{address: myAddr, option: group.VOTE_OPTION_YES, weight: "1"},
				{address: myOtherAddr, option: group.VOTE_OPTION_NO, weight: "2"},
			},
			policy: mockDecisionPolicy{
				AllowFn: func(tallyResult group.TallyResult, totalPower string) (group.DecisionPolicyResult, error) {
					return group.DecisionPolicyResult{Allow: false, Final: true}, nil
				},
			},
			expStatus:       group.PROPOSAL_STATUS_REJECTED,
			expVotesCleared: true,
			expEvents:       noEventsFn,
		},
		"proposal in flight": {
			votes: []memberVote{
				{address: myAddr, option: group.VOTE_OPTION_YES, weight: "1"},
				{address: myOtherAddr, option: group.VOTE_OPTION_NO, weight: "1"},
			},
			policy: mockDecisionPolicy{
				AllowFn: func(tallyResult group.TallyResult, totalPower string) (group.DecisionPolicyResult, error) {
					return group.DecisionPolicyResult{Allow: false, Final: false}, nil
				},
			},
			expStatus:       group.PROPOSAL_STATUS_SUBMITTED,
			expVotesCleared: false,
			expEvents:       noEventsFn,
		},
		"policy errors": {
			votes: []memberVote{
				{address: myAddr, option: group.VOTE_OPTION_YES, weight: "1"},
				{address: myOtherAddr, option: group.VOTE_OPTION_NO, weight: "2"},
			},
			policy: mockDecisionPolicy{
				AllowFn: func(tallyResult group.TallyResult, totalPower string) (group.DecisionPolicyResult, error) {
					return group.DecisionPolicyResult{}, errors.New("my test error")
				},
			},
			expStatus:       group.PROPOSAL_STATUS_REJECTED,
			expVotesCleared: true,
			expEvents: func(proposalID uint64) sdk.Events {
				return sdk.Events{
					sdk.NewEvent("cosmos.group.v1.EventTallyError",
						sdk.Attribute{Key: "error_message", Value: `"my test error"`},
						sdk.Attribute{Key: "proposal_id", Value: fmt.Sprintf(`"%d"`, proposalID)},
					),
				}
			},
		},
	}
	var (
		groupID    uint64
		proposalId uint64
	)
	for name, spec := range specs {
		groupID++
		proposalId++
		t.Run(name, func(t *testing.T) {
			em := sdk.NewEventManager()
			ctx := testCtx.Ctx.WithEventManager(em)
			totalWeight, err := math.NewDecFromString("0")
			require.NoError(t, err)
			// given a group, policy and persisted votes
			for _, v := range spec.votes {
				err := groupKeeper.groupMemberTable.Create(env.KVStoreService.OpenKVStore(ctx), &group.GroupMember{
					GroupId: groupID,
					Member:  &group.Member{Address: v.address, Weight: v.weight},
				})
				require.NoError(t, err)
				err = groupKeeper.voteTable.Create(env.KVStoreService.OpenKVStore(ctx), &group.Vote{
					ProposalId: proposalId,
					Voter:      v.address,
					Option:     v.option,
				})
				require.NoError(t, err)
			}
			myGroupInfo := group.GroupInfo{
				TotalWeight: totalWeight.String(),
			}
			myPolicy := group.GroupPolicyInfo{GroupId: groupID}
			err = myPolicy.SetDecisionPolicy(spec.policy)
			require.NoError(t, err)

			myProposal := &group.Proposal{
				Id:              proposalId,
				Status:          group.PROPOSAL_STATUS_SUBMITTED,
				VotingPeriodEnd: ctx.BlockTime().Add(time.Hour),
			}

			// when
			gotErr := groupKeeper.doTallyAndUpdate(ctx, myProposal, myGroupInfo, myPolicy)
			// then
			require.NoError(t, gotErr)
			assert.Equal(t, spec.expStatus, myProposal.Status)
			require.Equal(t, spec.expEvents(proposalId), em.Events())
			// and persistent state updated
			persistedVotes, err := groupKeeper.votesByProposal(ctx, groupID)
			require.NoError(t, err)
			if spec.expVotesCleared {
				assert.Empty(t, persistedVotes)
			} else {
				assert.Len(t, persistedVotes, len(spec.votes))
			}
		})
	}
}

var _ group.AccountKeeper = &mockAccountKeeper{}

// mockAccountKeeper is a mock implementation of the AccountKeeper interface for testing purposes.
type mockAccountKeeper struct {
	AddressCodecFn func() coreaddress.Codec
}

func (m mockAccountKeeper) AddressCodec() coreaddress.Codec {
	if m.AddressCodecFn == nil {
		panic("not expected to be called")
	}
	return m.AddressCodecFn()
}

func (m mockAccountKeeper) NewAccount(ctx context.Context, i sdk.AccountI) sdk.AccountI {
	panic("not expected to be called")
}

func (m mockAccountKeeper) GetAccount(ctx context.Context, address sdk.AccAddress) sdk.AccountI {
	panic("not expected to be called")
}

func (m mockAccountKeeper) SetAccount(ctx context.Context, i sdk.AccountI) {
	panic("not expected to be called")
}

func (m mockAccountKeeper) RemoveAccount(ctx context.Context, acc sdk.AccountI) {
	panic("not expected to be called")
}

// mockDecisionPolicy is a mock implementation of a decision policy for testing purposes.
type mockDecisionPolicy struct {
	fakeProtoType
	AllowFn func(tallyResult group.TallyResult, totalPower string) (group.DecisionPolicyResult, error)
}

func (m mockDecisionPolicy) Allow(tallyResult group.TallyResult, totalPower string) (group.DecisionPolicyResult, error) {
	if m.AllowFn == nil {
		panic("not expected to be called")
	}
	return m.AllowFn(tallyResult, totalPower)
}

func (m mockDecisionPolicy) GetVotingPeriod() time.Duration {
	panic("not expected to be called")
}

func (m mockDecisionPolicy) GetMinExecutionPeriod() time.Duration {
	panic("not expected to be called")
}

func (m mockDecisionPolicy) ValidateBasic() error {
	panic("not expected to be called")
}

func (m mockDecisionPolicy) Validate(g group.GroupInfo, config group.Config) error {
	panic("not expected to be called")
}

var (
	_ proto.Marshaler = (*fakeProtoType)(nil)
	_ proto.Message   = (*fakeProtoType)(nil)
)

// fakeProtoType is a struct used for mocking and testing purposes.
// Custom types can be converted into Any and back via internal CachedValue only.
type fakeProtoType struct{}

func (a fakeProtoType) Reset() {}

func (a fakeProtoType) String() string {
	return "testing"
}

func (a fakeProtoType) Marshal() ([]byte, error) {
	return nil, nil
}

func (a fakeProtoType) ProtoMessage() {}
