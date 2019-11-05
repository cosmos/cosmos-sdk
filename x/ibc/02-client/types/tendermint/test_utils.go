package tendermint

import (
	"github.com/tendermint/tendermint/crypto/tmhash"
	tmtypes "github.com/tendermint/tendermint/types"
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
