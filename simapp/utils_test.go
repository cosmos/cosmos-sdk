package simapp

import (
	"encoding/binary"
	"fmt"
	"time"

	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/tendermint/tendermint/crypto/ed25519"
	cmn "github.com/tendermint/tendermint/libs/common"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	delPk1   = ed25519.GenPrivKey().PubKey()
	delAddr1 = sdk.AccAddress(delPk1.Address())
)

func makeTestCodec() (cdc *codec.Codec) {
	cdc = codec.New()
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	authtypes.RegisterCodec(cdc)
	gov.RegisterCodec(cdc)
	return
}

func TestDecodeAccountStore(t *testing.T) {
	cdc := makeTestCodec()
	acc := authtypes.NewBaseAccountWithAddress(delAddr1)
	bz := cdc.MustMarshalBinaryBare(acc)

	require.Equal(t, fmt.Sprintf("%v\n%v", acc, acc), decodeAccountStore(cdc, cdc, bz, bz))
}

func TestDecodeMintStore(t *testing.T) {
	cdc := makeTestCodec()
	minter := minttypes.NewMinter(sdk.OneDec(), sdk.NewDec(15))
	bz := cdc.MustMarshalBinaryLengthPrefixed(minter)
	require.Equal(t, fmt.Sprintf("%v\n%v", minter, minter), decodeMintStore(cdc, cdc, bz, bz))
}

// func TestDecodeADistributiontore(t *testing.T) {

// }

// func TestDecodeStakingStore(t *testing.T) {

// }

func TestDecodeGovStore(t *testing.T) {
	cdc := makeTestCodec()

	endTime := time.Now().UTC()

	content := gov.ContentFromProposalType("test", "test", gov.ProposalTypeText)
	proposal := gov.NewProposal(content, 1, endTime, endTime.Add(24*time.Hour))
	proposalIDBz := make([]byte, 8)
	binary.LittleEndian.PutUint64(proposalIDBz, 1)
	deposit := gov.NewDeposit(1, delAddr1, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.OneInt())))
	vote := gov.NewVote(1, delAddr1, gov.OptionYes)

	kvPairs := cmn.KVPairs{
		cmn.KVPair{Key: gov.ProposalKey(1), Value: cdc.MustMarshalBinaryLengthPrefixed(proposal)},
		cmn.KVPair{Key: gov.InactiveProposalQueueKey(1, endTime), Value: proposalIDBz},
		cmn.KVPair{Key: gov.DepositKey(1, delAddr1), Value: cdc.MustMarshalBinaryLengthPrefixed(deposit)},
		cmn.KVPair{Key: gov.VoteKey(1, delAddr1), Value: cdc.MustMarshalBinaryLengthPrefixed(vote)},
		cmn.KVPair{Key: []byte{0x99}, Value: []byte{0x99}},
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
		t.Run(tt.name, func(t *testing.T) {
			switch i {
			case len(tests) - 1:
				require.Panics(t, func() { decodeGovStore(cdc, cdc, kvPairs[i], kvPairs[i]) })
			default:
				require.Equal(t, tt.expectedLog, decodeGovStore(cdc, cdc, kvPairs[i], kvPairs[i]))
			}
		})
	}
}
