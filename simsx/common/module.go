package common

import (
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"iter"
)

// WeightSource interface for retrieving weights based on a name and a default value.
type WeightSource interface {
	Get(name string, defaultValue uint32) uint32
}

// WeightedProposalMsgIter iterator for weighted gov proposal payload messages
type WeightedProposalMsgIter = iter.Seq2[uint32, FactoryMethod]

type (
	HasWeightedOperationsX interface {
		WeightedOperationsX(weight WeightSource, reg Registry)
	}
	HasWeightedOperationsXWithProposals interface {
		WeightedOperationsX(weights WeightSource, reg Registry, proposals WeightedProposalMsgIter,
			legacyProposals []simtypes.WeightedProposalContent) //nolint: staticcheck // used for legacy proposal types
	}
	HasProposalMsgsX interface {
		ProposalMsgsX(weights WeightSource, reg Registry)
	}
)
