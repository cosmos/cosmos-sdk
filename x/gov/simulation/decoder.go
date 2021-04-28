package simulation

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/kv"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding gov type.
func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key[:1], types.ProposalsKeyPrefix):
			var proposalA types.Proposal
			err := cdc.Unmarshal(kvA.Value, &proposalA)
			if err != nil {
				panic(err)
			}
			var proposalB types.Proposal
			err = cdc.Unmarshal(kvB.Value, &proposalB)
			if err != nil {
				panic(err)
			}
			return fmt.Sprintf("%v\n%v", proposalA, proposalB)

		case bytes.Equal(kvA.Key[:1], types.ActiveProposalQueuePrefix),
			bytes.Equal(kvA.Key[:1], types.InactiveProposalQueuePrefix),
			bytes.Equal(kvA.Key[:1], types.ProposalIDKey):
			proposalIDA := binary.LittleEndian.Uint64(kvA.Value)
			proposalIDB := binary.LittleEndian.Uint64(kvB.Value)
			return fmt.Sprintf("proposalIDA: %d\nProposalIDB: %d", proposalIDA, proposalIDB)

		case bytes.Equal(kvA.Key[:1], types.DepositsKeyPrefix):
			var depositA, depositB types.Deposit
			cdc.MustUnmarshal(kvA.Value, &depositA)
			cdc.MustUnmarshal(kvB.Value, &depositB)
			return fmt.Sprintf("%v\n%v", depositA, depositB)

		case bytes.Equal(kvA.Key[:1], types.VotesKeyPrefix):
			var voteA, voteB types.Vote
			cdc.MustUnmarshal(kvA.Value, &voteA)
			cdc.MustUnmarshal(kvB.Value, &voteB)
			return fmt.Sprintf("%v\n%v", voteA, voteB)

		default:
			panic(fmt.Sprintf("invalid governance key prefix %X", kvA.Key[:1]))
		}
	}
}
