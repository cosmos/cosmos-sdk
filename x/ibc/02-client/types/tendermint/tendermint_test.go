package tendermint

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/tendermint/tendermint/crypto/tmhash"
	tmtypes "github.com/tendermint/tendermint/types"
	"github.com/tendermint/tendermint/version"

	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

type TendermintTestSuite struct {
	suite.Suite

	privVal tmtypes.PrivValidator
	valSet  *tmtypes.ValidatorSet
	header  Header
	cs      ConsensusState
}

func (suite *TendermintTestSuite) SetupTest() {
	privVal := tmtypes.NewMockPV()
	val := tmtypes.NewValidator(privVal.GetPubKey(), 10)
	valSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{val})
	vsetHash := valSet.Hash()
	timestamp := time.Date(math.MaxInt64, 0, 0, 0, 0, 0, math.MaxInt64, time.UTC)
	tmHeader := tmtypes.Header{
		Version:            version.Consensus{Block: 2, App: 2},
		ChainID:            "mychain",
		Height:             3,
		Time:               timestamp,
		NumTxs:             100,
		TotalTxs:           1000,
		LastBlockID:        makeBlockID(make([]byte, tmhash.Size), math.MaxInt64, make([]byte, tmhash.Size)),
		LastCommitHash:     tmhash.Sum([]byte("last_commit_hash")),
		DataHash:           tmhash.Sum([]byte("data_hash")),
		ValidatorsHash:     vsetHash,
		NextValidatorsHash: vsetHash,
		ConsensusHash:      tmhash.Sum([]byte("consensus_hash")),
		AppHash:            tmhash.Sum([]byte("app_hash")),
		LastResultsHash:    tmhash.Sum([]byte("last_results_hash")),
		EvidenceHash:       tmhash.Sum([]byte("evidence_hash")),
		ProposerAddress:    privVal.GetPubKey().Address(),
	}
	hhash := tmHeader.Hash()
	blockID := makeBlockID(hhash, 3, tmhash.Sum([]byte("part_set")))
	voteSet := tmtypes.NewVoteSet("mychain", 3, 1, tmtypes.PrecommitType, valSet)
	commit, err := tmtypes.MakeCommit(blockID, 3, 1, voteSet, []tmtypes.PrivValidator{privVal})
	if err != nil {
		panic(err)
	}

	signedHeader := tmtypes.SignedHeader{
		Header: &tmHeader,
		Commit: commit,
	}

	header := Header{
		SignedHeader:     signedHeader,
		ValidatorSet:     valSet,
		NextValidatorSet: valSet,
	}

	root := commitment.NewRoot(tmhash.Sum([]byte("my root")))

	cs := ConsensusState{
		ChainID:          "mychain",
		Height:           3,
		Root:             root,
		NextValidatorSet: valSet,
	}

	// set fields in suite
	suite.privVal = privVal
	suite.valSet = valSet
	suite.header = header
	suite.cs = cs
}

func TestTendermintTestSuite(t *testing.T) {
	suite.Run(t, new(TendermintTestSuite))
}
