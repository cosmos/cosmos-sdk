package types

import (
	"bytes"
	"time"

	"github.com/tendermint/tendermint/crypto/tmhash"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"github.com/tendermint/tendermint/proto/tendermint/version"
	tmtypes "github.com/tendermint/tendermint/types"
)

// MakeBlockID is a copied unimported test function from tmtypes to use here
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
func CreateTestHeader(chainID string, height, trustedHeight int64, timestamp time.Time, tmValSet, tmTrustedVals *tmtypes.ValidatorSet, signers []tmtypes.PrivValidator) Header {
	var (
		valSet      *tmproto.ValidatorSet
		trustedVals *tmproto.ValidatorSet
	)
	vsetHash := tmValSet.Hash()
	tmHeader := tmtypes.Header{
		Version:            version.Consensus{Block: 2, App: 2},
		ChainID:            chainID,
		Height:             height,
		Time:               timestamp,
		LastBlockID:        MakeBlockID(make([]byte, tmhash.Size), 10_000, make([]byte, tmhash.Size)),
		LastCommitHash:     tmhash.Sum([]byte("last_commit_hash")),
		DataHash:           tmhash.Sum([]byte("data_hash")),
		ValidatorsHash:     vsetHash,
		NextValidatorsHash: vsetHash,
		ConsensusHash:      tmhash.Sum([]byte("consensus_hash")),
		AppHash:            tmhash.Sum([]byte("app_hash")),
		LastResultsHash:    tmhash.Sum([]byte("last_results_hash")),
		EvidenceHash:       tmhash.Sum([]byte("evidence_hash")),
		ProposerAddress:    tmValSet.Proposer.Address,
	}

	hhash := tmHeader.Hash()
	blockID := MakeBlockID(hhash, 3, tmhash.Sum([]byte("part_set")))
	voteSet := tmtypes.NewVoteSet(chainID, height, 1, tmproto.PrecommitType, tmValSet)
	commit, err := tmtypes.MakeCommit(blockID, height, 1, voteSet, signers, timestamp)
	if err != nil {
		panic(err)
	}

	signedHeader := tmproto.SignedHeader{
		Header: tmHeader.ToProto(),
		Commit: commit.ToProto(),
	}

	if tmValSet != nil {
		valSet, err = tmValSet.ToProto()
		if err != nil {
			panic(err)
		}
	}

	if tmTrustedVals != nil {
		trustedVals, err = tmTrustedVals.ToProto()
		if err != nil {
			panic(err)
		}
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

	switch {
	case altVal.VotingPower > suiteVal.VotingPower:
		return []tmtypes.PrivValidator{altPrivVal, suitePrivVal}
	case altVal.VotingPower < suiteVal.VotingPower:
		return []tmtypes.PrivValidator{suitePrivVal, altPrivVal}
	default:
		if bytes.Compare(altVal.Address, suiteVal.Address) == -1 {
			return []tmtypes.PrivValidator{altPrivVal, suitePrivVal}
		}
		return []tmtypes.PrivValidator{suitePrivVal, altPrivVal}
	}
}
