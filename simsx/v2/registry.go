package v2

import (
	"math/rand"
	"time"

	"github.com/huandu/skiplist"

	"github.com/cosmos/cosmos-sdk/simsx"
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
func (l *FutureOpsRegistry) Add(blockTime time.Time, fx simsx.SimMsgFactoryX) {
	if fx == nil {
		panic("message factory must not be nil")
	}
	if blockTime.IsZero() {
		return
	}
	var scheduledOps []simsx.SimMsgFactoryX
	if e := l.list.Get(blockTime); e != nil {
		scheduledOps = e.Value.([]simsx.SimMsgFactoryX)
	}
	scheduledOps = append(scheduledOps, fx)
	l.list.Set(blockTime, scheduledOps)
}

// PopScheduledFor retrieves and removes all scheduled operations up to the specified block time from the registry.
func (l *FutureOpsRegistry) PopScheduledFor(blockTime time.Time) []simsx.SimMsgFactoryX {
	var r []simsx.SimMsgFactoryX
	for {
		e := l.list.Front()
		if e == nil || e.Key().(time.Time).After(blockTime) {
			break
		}
		r = append(r, e.Value.([]simsx.SimMsgFactoryX)...)
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
