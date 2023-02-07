package simulation

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/kv"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding gov type.
func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key[:1], types.ProposalsKeyPrefix):
			var (
				proposalA v1beta1.Proposal
				proposalB v1beta1.Proposal

				proposalD v1.Proposal
				proposalC v1.Proposal
			)
			if err := cdc.Unmarshal(kvA.Value, &proposalC); err != nil {
				cdc.MustUnmarshal(kvA.Value, &proposalA)
			}

			if err := cdc.Unmarshal(kvB.Value, &proposalD); err != nil {
				cdc.MustUnmarshal(kvB.Value, &proposalB)
			}

			// this is to check if the proposal has been unmarshalled as v1 correctly (and not v1beta1)
			if proposalC.Title != "" || proposalD.Title != "" {
				return fmt.Sprintf("%v\n%v", proposalC, proposalD)
			}

			return fmt.Sprintf("%v\n%v", proposalA, proposalB)
		case bytes.Equal(kvA.Key[:1], types.ActiveProposalQueuePrefix),
			bytes.Equal(kvA.Key[:1], types.InactiveProposalQueuePrefix),
			bytes.Equal(kvA.Key[:1], types.ProposalIDKey):
			proposalIDA := binary.LittleEndian.Uint64(kvA.Value)
			proposalIDB := binary.LittleEndian.Uint64(kvB.Value)
			return fmt.Sprintf("proposalIDA: %d\nProposalIDB: %d", proposalIDA, proposalIDB)

		case bytes.Equal(kvA.Key[:1], types.DepositsKeyPrefix):
			var depositA, depositB v1beta1.Deposit
			cdc.MustUnmarshal(kvA.Value, &depositA)
			cdc.MustUnmarshal(kvB.Value, &depositB)
			return fmt.Sprintf("%v\n%v", depositA, depositB)

		case bytes.Equal(kvA.Key[:1], types.VotesKeyPrefix):
			var voteA, voteB v1beta1.Vote
			cdc.MustUnmarshal(kvA.Value, &voteA)
			cdc.MustUnmarshal(kvB.Value, &voteB)
			return fmt.Sprintf("%v\n%v", voteA, voteB)

		default:
			panic(fmt.Sprintf("invalid governance key prefix %X", kvA.Key[:1]))
		}
	}
}
