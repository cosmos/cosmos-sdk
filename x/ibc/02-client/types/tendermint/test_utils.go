package tendermint

import (
	"math"
	"time"

	"github.com/tendermint/tendermint/crypto/tmhash"
	tmtypes "github.com/tendermint/tendermint/types"
	"github.com/tendermint/tendermint/version"
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

func MakeHeader(height int64, valSet *tmtypes.ValidatorSet, nextValSet *tmtypes.ValidatorSet, signers []tmtypes.PrivValidator) Header {
	vsetHash := valSet.Hash()
	nextHash := nextValSet.Hash()
	timestamp := time.Date(math.MaxInt64, 0, 0, 0, 0, 0, math.MaxInt64, time.UTC)
	tmHeader := tmtypes.Header{
		Version:            version.Consensus{Block: 2, App: 2},
		ChainID:            "gaia",
		Height:             height,
		Time:               timestamp,
		NumTxs:             100,
		TotalTxs:           1000,
		LastBlockID:        makeBlockID(make([]byte, tmhash.Size), math.MaxInt64, make([]byte, tmhash.Size)),
		LastCommitHash:     tmhash.Sum([]byte("last_commit_hash")),
		DataHash:           tmhash.Sum([]byte("data_hash")),
		ValidatorsHash:     vsetHash,
		NextValidatorsHash: nextHash,
		ConsensusHash:      tmhash.Sum([]byte("consensus_hash")),
		AppHash:            tmhash.Sum([]byte("app_hash")),
		LastResultsHash:    tmhash.Sum([]byte("last_results_hash")),
		EvidenceHash:       tmhash.Sum([]byte("evidence_hash")),
		ProposerAddress:    signers[0].GetPubKey().Address(),
	}
	hhash := tmHeader.Hash()
	blockID := makeBlockID(hhash, 3, tmhash.Sum([]byte("part_set")))
	voteSet := tmtypes.NewVoteSet("gaia", height, 1, tmtypes.PrecommitType, valSet)
	commit, err := tmtypes.MakeCommit(blockID, height, 1, voteSet, signers)
	if err != nil {
		panic(err)
	}

	signedHeader := tmtypes.SignedHeader{
		Header: &tmHeader,
		Commit: commit,
	}

	return Header{
		SignedHeader:     signedHeader,
		ValidatorSet:     valSet,
		NextValidatorSet: nextValSet,
	}
}
