package types

import (
	"bytes"
	"math"
	"time"

	"github.com/tendermint/tendermint/crypto/tmhash"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"github.com/tendermint/tendermint/proto/tendermint/version"
	tmtypes "github.com/tendermint/tendermint/types"
)

// Copied unimported test functions from tmtypes to use them here
func MakeBlockID(hash []byte, partSetSize uint32, partSetHash []byte) tmtypes.BlockID {
	return tmtypes.BlockID{
		Hash: hash,
		PartSetHeader: tmtypes.PartSetHeader{
			Total: partSetSize,
			Hash:  partSetHash,
		},
	}
}

// CreateTestHeader creates a mock header for testing only.
func CreateTestHeader(chainID string, height, trustedHeight int64, timestamp time.Time, valSet, trustedVals *tmtypes.ValidatorSet, signers []tmtypes.PrivValidator) Header {
	vsetHash := valSet.Hash()
	tmHeader := tmtypes.Header{
		Version:            version.Consensus{Block: 2, App: 2},
		ChainID:            chainID,
		Height:             height,
		Time:               timestamp,
		LastBlockID:        MakeBlockID(make([]byte, tmhash.Size), math.MaxUint32, make([]byte, tmhash.Size)),
		LastCommitHash:     tmhash.Sum([]byte("last_commit_hash")),
		DataHash:           tmhash.Sum([]byte("data_hash")),
		ValidatorsHash:     vsetHash,
		NextValidatorsHash: vsetHash,
		ConsensusHash:      tmhash.Sum([]byte("consensus_hash")),
		AppHash:            tmhash.Sum([]byte("app_hash")),
		LastResultsHash:    tmhash.Sum([]byte("last_results_hash")),
		EvidenceHash:       tmhash.Sum([]byte("evidence_hash")),
		ProposerAddress:    valSet.Proposer.Address,
	}

	hhash := tmHeader.Hash()
	blockID := MakeBlockID(hhash, 3, tmhash.Sum([]byte("part_set")))
	voteSet := tmtypes.NewVoteSet(chainID, height, 1, tmproto.PrecommitType, valSet)
	commit, err := tmtypes.MakeCommit(blockID, height, 1, voteSet, signers, timestamp)
	if err != nil {
		panic(err)
	}

	signedHeader := tmtypes.SignedHeader{
		Header: &tmHeader,
		Commit: commit,
	}

	return Header{
		SignedHeader:      signedHeader,
		ValidatorSet:      valSet,
		TrustedHeight:     uint64(trustedHeight),
		TrustedValidators: trustedVals,
	}
}

// CreateSortedSignerArray takes two PrivValidators, and the corresponding Validator structs
// (including voting power). It returns a signer array of PrivValidators that matches the
// sorting of ValidatorSet.
// The sorting is first by .VotingPower (descending), with secondary index of .Address (ascending).
func CreateSortedSignerArray(altPrivVal, suitePrivVal tmtypes.PrivValidator,
	altVal, suiteVal *tmtypes.Validator) []tmtypes.PrivValidator {
	var bothSigners []tmtypes.PrivValidator

	if altVal.VotingPower > suiteVal.VotingPower {
		bothSigners = []tmtypes.PrivValidator{altPrivVal, suitePrivVal}
	} else if altVal.VotingPower < suiteVal.VotingPower {
		bothSigners = []tmtypes.PrivValidator{suitePrivVal, altPrivVal}
	} else {
		if bytes.Compare(altVal.Address, suiteVal.Address) == -1 {
			bothSigners = []tmtypes.PrivValidator{altPrivVal, suitePrivVal}
		} else {
			bothSigners = []tmtypes.PrivValidator{suitePrivVal, altPrivVal}
		}
	}

	return bothSigners
}
