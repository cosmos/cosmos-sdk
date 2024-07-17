package simsx

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"golang.org/x/exp/maps"
	"iter"
	"math/rand"
	"slices"
	"strings"
)

var _ Registry = &SimsV2Reg{}

type SimsV2Reg map[string]WeightedFactory

func (s SimsV2Reg) Add(weight uint32, f SimMsgFactoryX) {
	msgType := f.MsgType()
	msgTypeURL := sdk.MsgTypeURL(msgType)
	if _, exists := s[msgTypeURL]; exists {
		panic("type is already registered: " + msgTypeURL)
	}
	s[msgTypeURL] = WeightedFactory{Weight: weight, Factory: f}
}

func (s SimsV2Reg) NextFactoryFn(r *rand.Rand) func() SimMsgFactoryX {
	factories := maps.Values(s)
	slices.SortFunc(factories, func(a, b WeightedFactory) int { // sort to make deterministic
		return strings.Compare(sdk.MsgTypeURL(a.Factory.MsgType()), sdk.MsgTypeURL(b.Factory.MsgType()))
	})
	factCount := len(factories)
	r.Shuffle(factCount, func(i, j int) {
		factories[i], factories[j] = factories[j], factories[i]
	})
	var totalWeight int
	for k := range factories {
		totalWeight += k
	}
	return func() SimMsgFactoryX {
		// this is copied from old sims WeightedOperations.getSelectOpFn
		x := r.Intn(totalWeight)
		for i := 0; i < factCount; i++ {
			if x <= int(factories[i].Weight) {
				return factories[i].Factory
			}
			x -= int(factories[i].Weight)
		}
		// shouldn't happen
		return factories[0].Factory
	}
}

func (s SimsV2Reg) Iterator() iter.Seq2[uint32, SimMsgFactoryX] {
	x := maps.Values(s)
	slices.SortFunc(x, func(a, b WeightedFactory) int {
		return a.Compare(b)
	})
	return func(yield func(uint32, SimMsgFactoryX) bool) {
		for _, v := range x {
			if !yield(v.Weight, v.Factory) {
				return
			}
		}
	}
}

type WeightedFactory struct {
	Weight  uint32
	Factory SimMsgFactoryX
}

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
