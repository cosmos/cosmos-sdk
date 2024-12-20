package v2

import (
	"github.com/cosmos/cosmos-sdk/simsx/common"
	runnercommon "github.com/cosmos/cosmos-sdk/simsx/runner/common"
	"math/rand"
	"time"

	"github.com/huandu/skiplist"
)

type (
	WeightedProposalMsgIter = common.WeightedProposalMsgIter
	WeightedFactory         = runnercommon.WeightedFactory
)

// FutureOpsRegistry is a registry for scheduling and retrieving operations mapped to future block times.
type FutureOpsRegistry struct {
	list *skiplist.SkipList
}

// NewFutureOpsRegistry constructor
func NewFutureOpsRegistry() *FutureOpsRegistry {
	list := skiplist.New(timeComparator{})
	list.SetRandSource(rand.NewSource(1))
	return &FutureOpsRegistry{list: list}
}

// Add schedules a new operation for the given block time
func (l *FutureOpsRegistry) Add(blockTime time.Time, fx common.SimMsgFactoryX) {
	if fx == nil {
		panic("message factory must not be nil")
	}
	if blockTime.IsZero() {
		return
	}
	var scheduledOps []common.SimMsgFactoryX
	if e := l.list.Get(blockTime); e != nil {
		scheduledOps = e.Value.([]common.SimMsgFactoryX)
	}
	scheduledOps = append(scheduledOps, fx)
	l.list.Set(blockTime, scheduledOps)
}

// PopScheduledFor retrieves and removes all scheduled operations up to the specified block time from the registry.
func (l *FutureOpsRegistry) PopScheduledFor(blockTime time.Time) []common.SimMsgFactoryX {
	var r []common.SimMsgFactoryX
	for {
		e := l.list.Front()
		if e == nil || e.Key().(time.Time).After(blockTime) {
			break
		}
		r = append(r, e.Value.([]common.SimMsgFactoryX)...)
		l.list.RemoveFront()
	}
	return r
}

var _ skiplist.Comparable = timeComparator{}

// used for SkipList
type timeComparator struct{}

func (t timeComparator) Compare(lhs, rhs interface{}) int {
	return lhs.(time.Time).Compare(rhs.(time.Time))
}

func (t timeComparator) CalcScore(key interface{}) float64 {
	return float64(key.(time.Time).UnixNano())
}

var _ common.Registry = &UnorderedRegistry{}

// UnorderedRegistry represents a collection of WeightedFactory elements without guaranteed order.
// It is used to maintain factories coupled with their associated weights for simulation purposes.
type UnorderedRegistry []WeightedFactory

func NewUnorderedRegistry() *UnorderedRegistry {
	r := make(UnorderedRegistry, 0)
	return &r
}

// Add appends a new WeightedFactory with the provided weight and factory to the UnorderedRegistry.
func (x *UnorderedRegistry) Add(weight uint32, f common.SimMsgFactoryX) {
	if weight == 0 {
		return
	}
	if f == nil {
		panic("message factory must not be nil")
	}
	*x = append(*x, WeightedFactory{Weight: weight, Factory: f})
}

// Elements returns all collected elements
func (x *UnorderedRegistry) Elements() []WeightedFactory {
	return *x
}
