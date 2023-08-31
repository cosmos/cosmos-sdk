package baseapp_test

import (
	"bytes"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/secp256k1"
	cmtprotocrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	protoio "github.com/cosmos/gogoproto/io"
	"github.com/cosmos/gogoproto/proto"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/baseapp/testutil/mock"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	chainID = "chain-id"
)

type testValidator struct {
	consAddr sdk.ConsAddress
	tmPk     cmtprotocrypto.PublicKey
	privKey  secp256k1.PrivKey
}

func newTestValidator() testValidator {
	privkey := secp256k1.GenPrivKey()
	pubkey := privkey.PubKey()
	tmPk := cmtprotocrypto.PublicKey{
		Sum: &cmtprotocrypto.PublicKey_Secp256K1{
			Secp256K1: pubkey.Bytes(),
		},
	}

	return testValidator{
		consAddr: sdk.ConsAddress(pubkey.Address()),
		tmPk:     tmPk,
		privKey:  privkey,
	}
}

func (t testValidator) toValidator(power int64) abci.Validator {
	return abci.Validator{
		Address: t.consAddr.Bytes(),
		Power:   power,
	}
}

type ABCIUtilsTestSuite struct {
	suite.Suite

	valStore *mock.MockValidatorStore
	vals     [3]testValidator
	ctx      sdk.Context
}

func NewABCIUtilsTestSuite(t *testing.T) *ABCIUtilsTestSuite {
	t.Helper()
	// create 3 validators
	s := &ABCIUtilsTestSuite{
		vals: [3]testValidator{
			newTestValidator(),
			newTestValidator(),
			newTestValidator(),
		},
	}

	// create mock
	ctrl := gomock.NewController(t)
	valStore := mock.NewMockValidatorStore(ctrl)
	s.valStore = valStore

	// set up mock
	s.valStore.EXPECT().GetPubKeyByConsAddr(gomock.Any(), s.vals[0].consAddr.Bytes()).Return(s.vals[0].tmPk, nil).AnyTimes()
	s.valStore.EXPECT().GetPubKeyByConsAddr(gomock.Any(), s.vals[1].consAddr.Bytes()).Return(s.vals[1].tmPk, nil).AnyTimes()
	s.valStore.EXPECT().GetPubKeyByConsAddr(gomock.Any(), s.vals[2].consAddr.Bytes()).Return(s.vals[2].tmPk, nil).AnyTimes()

	// create context
	s.ctx = sdk.Context{}.WithConsensusParams(cmtproto.ConsensusParams{
		Abci: &cmtproto.ABCIParams{
			VoteExtensionsEnableHeight: 2,
		},
	})
	return s
}

func TestABCIUtilsTestSuite(t *testing.T) {
	suite.Run(t, NewABCIUtilsTestSuite(t))
}

// check ValidateVoteExtensions works when all nodes have CommitBlockID votes
func (s *ABCIUtilsTestSuite) TestValidateVoteExtensionsHappyPath() {
	ext := []byte("vote-extension")
	cve := cmtproto.CanonicalVoteExtension{
		Extension: ext,
		Height:    2,
		Round:     int64(0),
		ChainId:   chainID,
	}

	bz, err := marshalDelimitedFn(&cve)
	s.Require().NoError(err)

	extSig0, err := s.vals[0].privKey.Sign(bz)
	s.Require().NoError(err)

	extSig1, err := s.vals[1].privKey.Sign(bz)
	s.Require().NoError(err)

	extSig2, err := s.vals[2].privKey.Sign(bz)
	s.Require().NoError(err)

	llc := abci.ExtendedCommitInfo{
		Round: 0,
		Votes: []abci.ExtendedVoteInfo{
			{
				Validator:          s.vals[0].toValidator(333),
				VoteExtension:      ext,
				ExtensionSignature: extSig0,
				BlockIdFlag:        cmtproto.BlockIDFlagCommit,
			},
			{
				Validator:          s.vals[1].toValidator(333),
				VoteExtension:      ext,
				ExtensionSignature: extSig1,
				BlockIdFlag:        cmtproto.BlockIDFlagCommit,
			},
			{
				Validator:          s.vals[2].toValidator(334),
				VoteExtension:      ext,
				ExtensionSignature: extSig2,
				BlockIdFlag:        cmtproto.BlockIDFlagCommit,
			},
		},
	}
	// expect-pass (votes of height 2 are included in next block)
	s.Require().NoError(baseapp.ValidateVoteExtensions(s.ctx, s.valStore, 3, chainID, llc))
}

// check ValidateVoteExtensions works when a single node has submitted a BlockID_Absent
func (s *ABCIUtilsTestSuite) TestValidateVoteExtensionsSingleVoteAbsent() {
	ext := []byte("vote-extension")
	cve := cmtproto.CanonicalVoteExtension{
		Extension: ext,
		Height:    2,
		Round:     int64(0),
		ChainId:   chainID,
	}

	bz, err := marshalDelimitedFn(&cve)
	s.Require().NoError(err)

	extSig0, err := s.vals[0].privKey.Sign(bz)
	s.Require().NoError(err)

	extSig2, err := s.vals[2].privKey.Sign(bz)
	s.Require().NoError(err)

	llc := abci.ExtendedCommitInfo{
		Round: 0,
		Votes: []abci.ExtendedVoteInfo{
			{
				Validator:          s.vals[0].toValidator(333),
				VoteExtension:      ext,
				ExtensionSignature: extSig0,
				BlockIdFlag:        cmtproto.BlockIDFlagCommit,
			},
			// validator of power <1/3 is missing, so commit-info shld still be valid
			{
				Validator:   s.vals[1].toValidator(333),
				BlockIdFlag: cmtproto.BlockIDFlagAbsent,
			},
			{
				Validator:          s.vals[2].toValidator(334),
				VoteExtension:      ext,
				ExtensionSignature: extSig2,
				BlockIdFlag:        cmtproto.BlockIDFlagCommit,
			},
		},
	}
	// expect-pass (votes of height 2 are included in next block)
	s.Require().NoError(baseapp.ValidateVoteExtensions(s.ctx, s.valStore, 3, chainID, llc))
}

// check ValidateVoteExtensions works when a single node has submitted a BlockID_Nil
func (s *ABCIUtilsTestSuite) TestValidateVoteExtensionsSingleVoteNil() {
	ext := []byte("vote-extension")
	cve := cmtproto.CanonicalVoteExtension{
		Extension: ext,
		Height:    2,
		Round:     int64(0),
		ChainId:   chainID,
	}

	bz, err := marshalDelimitedFn(&cve)
	s.Require().NoError(err)

	extSig0, err := s.vals[0].privKey.Sign(bz)
	s.Require().NoError(err)

	extSig2, err := s.vals[2].privKey.Sign(bz)
	s.Require().NoError(err)

	llc := abci.ExtendedCommitInfo{
		Round: 0,
		Votes: []abci.ExtendedVoteInfo{
			{
				Validator:          s.vals[0].toValidator(333),
				VoteExtension:      ext,
				ExtensionSignature: extSig0,
				BlockIdFlag:        cmtproto.BlockIDFlagCommit,
			},
			// validator of power <1/3 is missing, so commit-info should still be valid
			{
				Validator:   s.vals[1].toValidator(333),
				BlockIdFlag: cmtproto.BlockIDFlagNil,
			},
			{
				Validator:          s.vals[2].toValidator(334),
				VoteExtension:      ext,
				ExtensionSignature: extSig2,
				BlockIdFlag:        cmtproto.BlockIDFlagCommit,
			},
		},
	}
	// expect-pass (votes of height 2 are included in next block)
	s.Require().NoError(baseapp.ValidateVoteExtensions(s.ctx, s.valStore, 3, chainID, llc))
}

// check ValidateVoteExtensions works when two nodes have submitted a BlockID_Nil / BlockID_Absent
func (s *ABCIUtilsTestSuite) TestValidateVoteExtensionsTwoVotesNilAbsent() {
	ext := []byte("vote-extension")
	cve := cmtproto.CanonicalVoteExtension{
		Extension: ext,
		Height:    2,
		Round:     int64(0),
		ChainId:   chainID,
	}

	bz, err := marshalDelimitedFn(&cve)
	s.Require().NoError(err)

	extSig0, err := s.vals[0].privKey.Sign(bz)
	s.Require().NoError(err)

	llc := abci.ExtendedCommitInfo{
		Round: 0,
		Votes: []abci.ExtendedVoteInfo{
			// validator of power >2/3 is missing, so commit-info should not be valid
			{
				Validator:          s.vals[0].toValidator(333),
				BlockIdFlag:        cmtproto.BlockIDFlagCommit,
				VoteExtension:      ext,
				ExtensionSignature: extSig0,
			},
			{
				Validator:   s.vals[1].toValidator(333),
				BlockIdFlag: cmtproto.BlockIDFlagNil,
			},
			{
				Validator:     s.vals[2].toValidator(334),
				VoteExtension: ext,
				BlockIdFlag:   cmtproto.BlockIDFlagAbsent,
			},
		},
	}

	// expect-pass (votes of height 2 are included in next block)
	s.Require().Error(baseapp.ValidateVoteExtensions(s.ctx, s.valStore, 3, chainID, llc))
}

func marshalDelimitedFn(msg proto.Message) ([]byte, error) {
	var buf bytes.Buffer
	if err := protoio.NewDelimitedWriter(&buf).WriteMsg(msg); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
