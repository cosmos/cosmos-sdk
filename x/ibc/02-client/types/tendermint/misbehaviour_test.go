package tendermint

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/crypto/tmhash"
	tmtypes "github.com/tendermint/tendermint/types"
	yaml "gopkg.in/yaml.v2"
)

// Copied unimported test functions from tmtypes to use them here
func makeBlockID(hash []byte, partSetSize int, partSetHash []byte) tmtypes.BlockID {
	return tmtypes.BlockID{
		Hash: hash,
		PartsHeader: tmtypes.PartSetHeader{
			Total: partSetSize,
			Hash:  partSetHash,
		},
	}

}

func makeVote(val tmtypes.PrivValidator, chainID string, valIndex int, height int64, round, step int, blockID tmtypes.BlockID) *tmtypes.Vote {
	addr := val.GetPubKey().Address()
	v := &tmtypes.Vote{
		ValidatorAddress: addr,
		ValidatorIndex:   valIndex,
		Height:           height,
		Round:            round,
		Type:             tmtypes.SignedMsgType(step),
		BlockID:          blockID,
	}
	err := val.SignVote(chainID, v)
	if err != nil {
		panic(err)
	}
	return v
}

func randomDuplicatedVoteEvidence() *tmtypes.DuplicateVoteEvidence {
	val := tmtypes.NewMockPV()
	blockID := makeBlockID(tmhash.Sum([]byte("blockhash")), 1000, tmhash.Sum([]byte("partshash")))
	blockID2 := makeBlockID(tmhash.Sum([]byte("blockhash2")), 1000, tmhash.Sum([]byte("partshash")))
	const chainID = "mychain"
	return &tmtypes.DuplicateVoteEvidence{
		PubKey: val.GetPubKey(),
		VoteA:  makeVote(val, chainID, 0, 10, 2, 1, blockID),
		VoteB:  makeVote(val, chainID, 0, 10, 2, 1, blockID2),
	}
}

func TestString(t *testing.T) {
	dupEv := randomDuplicatedVoteEvidence()
	ev := Evidence{
		DuplicateVoteEvidence: dupEv,
		ChainID:               "mychain",
		ValidatorPower:        10,
		TotalPower:            50,
	}

	byteStr, err := yaml.Marshal(ev)
	require.Nil(t, err)
	require.Equal(t, string(byteStr), ev.String(), "Evidence String method does not work as expected")

}

func TestValidateBasic(t *testing.T) {
	dupEv := randomDuplicatedVoteEvidence()

	// good evidence
	ev := Evidence{
		DuplicateVoteEvidence: dupEv,
		ChainID:               "mychain",
		ValidatorPower:        10,
		TotalPower:            50,
	}

	err := ev.ValidateBasic()
	require.Nil(t, err, "good evidence failed on ValidateBasic: %v", err)

	// invalid duplicate evidence
	ev.DuplicateVoteEvidence.VoteA = nil
	err = ev.ValidateBasic()
	require.NotNil(t, err, "invalid duplicate evidence passed on ValidateBasic")

	// reset duplicate evidence to be valid, and set empty chainID
	dupEv = randomDuplicatedVoteEvidence()
	ev.DuplicateVoteEvidence = dupEv
	ev.ChainID = ""
	err = ev.ValidateBasic()
	require.NotNil(t, err, "invalid chain-id passed on ValidateBasic")

	// reset chainID and set 0 validator power
	ev.ChainID = "mychain"
	ev.ValidatorPower = 0
	err = ev.ValidateBasic()
	require.NotNil(t, err, "invalid validator power passed on ValidateBasic")

	// reset validator power and set invalid total power
	ev.ValidatorPower = 10
	ev.TotalPower = 9
	err = ev.ValidateBasic()
	require.NotNil(t, err, "invalid total power passed on ValidateBasic")

}
