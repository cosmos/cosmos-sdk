package simulation

import (
	"encoding/binary"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/crypto/ed25519"
	tmkv "github.com/tendermint/tendermint/libs/kv"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

var (
	delPk1   = ed25519.GenPrivKey().PubKey()
	delAddr1 = sdk.AccAddress(delPk1.Address())
)

func makeTestCodec() (cdc *codec.Codec) {
	cdc = codec.New()
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	types.RegisterCodec(cdc)
	return
}

func TestDecodeStore(t *testing.T) {
	cdc := makeTestCodec()

	endTime := time.Now().UTC()

	content := types.ContentFromProposalType("test", "test", types.ProposalTypeText)
	proposal := types.NewProposal(content, 1, endTime, endTime.Add(24*time.Hour))
	proposalIDBz := make([]byte, 8)
	binary.LittleEndian.PutUint64(proposalIDBz, 1)
	deposit := types.NewDeposit(1, delAddr1, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.OneInt())))
	vote := types.NewVote(1, delAddr1, types.OptionYes)

	kvPairs := tmkv.Pairs{
		tmkv.Pair{Key: types.ProposalKey(1), Value: cdc.MustMarshalBinaryLengthPrefixed(proposal)},
		tmkv.Pair{Key: types.InactiveProposalQueueKey(1, endTime), Value: proposalIDBz},
		tmkv.Pair{Key: types.DepositKey(1, delAddr1), Value: cdc.MustMarshalBinaryLengthPrefixed(deposit)},
		tmkv.Pair{Key: types.VoteKey(1, delAddr1), Value: cdc.MustMarshalBinaryLengthPrefixed(vote)},
		tmkv.Pair{Key: []byte{0x99}, Value: []byte{0x99}},
	}

	tests := []struct {
		name        string
		expectedLog string
	}{
		{"proposals", fmt.Sprintf("%v\n%v", proposal, proposal)},
		{"proposal IDs", "proposalIDA: 1\nProposalIDB: 1"},
		{"deposits", fmt.Sprintf("%v\n%v", deposit, deposit)},
		{"votes", fmt.Sprintf("%v\n%v", vote, vote)},
		{"other", ""},
	}

	for i, tt := range tests {
		i, tt := i, tt
		t.Run(tt.name, func(t *testing.T) {
			switch i {
			case len(tests) - 1:
				require.Panics(t, func() { DecodeStore(cdc, kvPairs[i], kvPairs[i]) }, tt.name)
			default:
				require.Equal(t, tt.expectedLog, DecodeStore(cdc, kvPairs[i], kvPairs[i]), tt.name)
			}
		})
	}
}
