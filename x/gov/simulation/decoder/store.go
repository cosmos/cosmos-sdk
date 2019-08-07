package decoder

import (
	"bytes"
	"encoding/binary"
	"fmt"

	cmn "github.com/tendermint/tendermint/libs/common"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

// DecodeStore unmarshals the KVPair's Value to the corresponding gov type
func DecodeStore(cdcA, cdcB *codec.Codec, kvA, kvB cmn.KVPair) string {
	switch {
	case bytes.Equal(kvA.Key[:1], types.ProposalsKeyPrefix):
		var proposalA, proposalB types.Proposal
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &proposalA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &proposalB)
		return fmt.Sprintf("%v\n%v", proposalA, proposalB)

	case bytes.Equal(kvA.Key[:1], types.ActiveProposalQueuePrefix),
		bytes.Equal(kvA.Key[:1], types.InactiveProposalQueuePrefix),
		bytes.Equal(kvA.Key[:1], types.ProposalIDKey):
		proposalIDA := binary.LittleEndian.Uint64(kvA.Value)
		proposalIDB := binary.LittleEndian.Uint64(kvB.Value)
		return fmt.Sprintf("proposalIDA: %d\nProposalIDB: %d", proposalIDA, proposalIDB)

	case bytes.Equal(kvA.Key[:1], types.DepositsKeyPrefix):
		var depositA, depositB types.Deposit
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &depositA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &depositB)
		return fmt.Sprintf("%v\n%v", depositA, depositB)

	case bytes.Equal(kvA.Key[:1], types.VotesKeyPrefix):
		var voteA, voteB types.Vote
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &voteA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &voteB)
		return fmt.Sprintf("%v\n%v", voteA, voteB)

	default:
		panic(fmt.Sprintf("invalid governance key prefix %X", kvA.Key[:1]))
	}
}
