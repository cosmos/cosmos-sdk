package common

import (
	"github.com/cosmos/cosmos-sdk/simsx/common"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"maps"
	"slices"
	"strings"
)

type WeightedProposalMsgIter = common.WeightedProposalMsgIter

var _ common.Registry = &UniqueTypeRegistry{}

type UniqueTypeRegistry map[string]WeightedFactory

func NewUniqueTypeRegistry() UniqueTypeRegistry {
	return make(UniqueTypeRegistry)
}

func (s UniqueTypeRegistry) Add(weight uint32, f common.SimMsgFactoryX) {
	if weight == 0 {
		return
	}
	if f == nil {
		panic("message factory must not be nil")
	}
	msgType := f.MsgType()
	msgTypeURL := sdk.MsgTypeURL(msgType)
	if _, exists := s[msgTypeURL]; exists {
		panic("type is already registered: " + msgTypeURL)
	}
	s[msgTypeURL] = WeightedFactory{Weight: weight, Factory: f}
}

// Iterator returns an iterator function for a Go for loop sorted by weight desc.
func (s UniqueTypeRegistry) Iterator() WeightedProposalMsgIter {
	sortedWeightedFactory := slices.SortedFunc(maps.Values(s), func(a, b WeightedFactory) int {
		return a.Compare(b)
	})

	return func(yield func(uint32, common.FactoryMethod) bool) {
		for _, v := range sortedWeightedFactory {
			if !yield(v.Weight, v.Factory.Create()) {
				return
			}
		}
	}
}

// WeightedFactory is a Weight Factory tuple
type WeightedFactory struct {
	Weight  uint32
	Factory common.SimMsgFactoryX
}

// Compare compares the WeightedFactory f with another WeightedFactory b.
// The comparison is done by comparing the weight of f with the weight of b.
// If the weight of f is greater than the weight of b, it returns 1.
// If the weight of f is less than the weight of b, it returns -1.
// If the weights are equal, it compares the TypeURL of the factories using strings.Compare.
// Returns an integer indicating the comparison result.
func (f WeightedFactory) Compare(b WeightedFactory) int {
	switch {
	case f.Weight > b.Weight:
		return 1
	case f.Weight < b.Weight:
		return -1
	default:
		return strings.Compare(sdk.MsgTypeURL(f.Factory.MsgType()), sdk.MsgTypeURL(b.Factory.MsgType()))
	}
}
