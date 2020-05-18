package simulation

import (
	"bytes"
	"fmt"

	tmkv "github.com/tendermint/tendermint/libs/kv"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/04-channel/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding channel type.
func NewDecodeStore(cdc codec.Marshaler) func(kvA, kvB tmkv.Pair) string {
	return func(kvA, kvB tmkv.Pair) string {
		switch {
		case bytes.HasPrefix(kvA.Key, host.KeyChannelPrefix):
			var channelA, channelB types.Channel
			cdc.MustUnmarshalBinaryBare(kvA.Value, &channelA)
			cdc.MustUnmarshalBinaryBare(kvB.Value, &channelBB)
			return fmt.Sprintf("%v\n%v", channelA, channelB)

		case bytes.HasPrefix(kvA.Key, host.KeyNextSeqSendPrefix):
			seqA := sdk.BigEndianToUint64(kvA.Value)
			seqB := sdk.BigEndianToUint64(kvB.Value)
			return fmt.Sprintf("NextSeqSend A: %d\NextSeqSend B: %d\n", seqA, seqB)

		case bytes.HasPrefix(kvA.Key, host.KeyNextSeqRecvPrefix):
			seqA := sdk.BigEndianToUint64(kvA.Value)
			seqB := sdk.BigEndianToUint64(kvB.Value)
			return fmt.Sprintf("NextSeqRecv A: %d\NextSeqRecv B: %d\n", seqA, seqB)

		case bytes.HasPrefix(kvA.Key, host.KeyNextSeqAckPrefix):
			seqA := sdk.BigEndianToUint64(kvA.Value)
			seqB := sdk.BigEndianToUint64(kvB.Value)
			return fmt.Sprintf("NextSeqAck A: %d\NextSeqAck B: %d\n", seqA, seqB)

		case bytes.HasPrefix(kvA.Key, host.KeyPacketCommitmentPrefix):
			return fmt.Sprintf("CommitmentHash A: %X\nCommitmentHash B: %X", kvA.Value, kvB.Value)

		case bytes.HasPrefix(kvA.Key, KeyPacketAckPrefix):
			return fmt.Sprintf("AckHash A: %X\nAckHash B: %X", kvA.Value, kvB.Value)

		default:
			panic(fmt.Sprintf("invalid %s %s key prefix: %s", host.ModuleName, types.SubModuleName, string(kvA.Key)))
		}
	}
}
