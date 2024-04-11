package cmtservice

import (
	"errors"
	fmt "fmt"

	tmcrypto "buf.build/gen/go/tendermint/tendermint/protocolbuffers/go/tendermint/crypto"
	tmtypes "buf.build/gen/go/tendermint/tendermint/protocolbuffers/go/tendermint/types"
	tmversion "buf.build/gen/go/tendermint/tendermint/protocolbuffers/go/tendermint/version"
	tmv1beta1 "cosmossdk.io/api/cosmos/base/tendermint/v1beta1"
	"cosmossdk.io/core/address"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/crypto/secp256k1"
	cmttypes "github.com/cometbft/cometbft/types"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func headerToProto(h *cmttypes.Header) *tmtypes.Header {
	return &tmtypes.Header{
		Version: &tmversion.Consensus{
			Block: h.Version.Block,
			App:   h.Version.App,
		},
		ChainId: h.ChainID,
		Height:  h.Height,
		Time:    timestamppb.New(h.Time),
		LastBlockId: &tmtypes.BlockID{
			Hash: h.LastBlockID.Hash,
			PartSetHeader: &tmtypes.PartSetHeader{
				Total: h.LastBlockID.PartSetHeader.Total,
				Hash:  h.LastBlockID.PartSetHeader.Hash,
			},
		},
		LastCommitHash:     h.LastCommitHash,
		DataHash:           h.DataHash,
		ValidatorsHash:     h.ValidatorsHash,
		NextValidatorsHash: h.NextValidatorsHash,
		ConsensusHash:      h.ConsensusHash,
		AppHash:            h.AppHash,
		LastResultsHash:    h.LastResultsHash,
		EvidenceHash:       h.EvidenceHash,
		ProposerAddress:    h.ProposerAddress,
	}

}

func evidenceToProto(evidence cmttypes.Evidence) (*tmtypes.Evidence, error) {
	if evidence == nil {
		return nil, errors.New("nil evidence")
	}

	switch evi := evidence.(type) {
	case *cmttypes.DuplicateVoteEvidence:
		pbev := &tmtypes.DuplicateVoteEvidence{
			VoteA: &tmtypes.Vote{
				Type:   tmtypes.SignedMsgType(evi.VoteA.Type),
				Height: evi.VoteA.Height,
				Round:  evi.VoteA.Round,
				BlockId: &tmtypes.BlockID{
					Hash: evi.VoteA.BlockID.Hash,
					PartSetHeader: &tmtypes.PartSetHeader{
						Total: evi.VoteA.BlockID.PartSetHeader.Total,
						Hash:  evi.VoteA.BlockID.PartSetHeader.Hash,
					},
				},
				Timestamp:          timestamppb.New(evi.VoteA.Timestamp),
				ValidatorAddress:   evi.VoteA.ValidatorAddress,
				ValidatorIndex:     evi.VoteA.ValidatorIndex,
				Signature:          evi.VoteA.Signature,
				Extension:          evi.VoteA.Extension,
				ExtensionSignature: evi.VoteA.ExtensionSignature,
			},
			VoteB: &tmtypes.Vote{
				Type:   tmtypes.SignedMsgType(evi.VoteB.Type),
				Height: evi.VoteB.Height,
				Round:  evi.VoteB.Round,
				BlockId: &tmtypes.BlockID{
					Hash: evi.VoteB.BlockID.Hash,
					PartSetHeader: &tmtypes.PartSetHeader{
						Total: evi.VoteB.BlockID.PartSetHeader.Total,
						Hash:  evi.VoteB.BlockID.PartSetHeader.Hash,
					},
				},
				Timestamp:          timestamppb.New(evi.VoteB.Timestamp),
				ValidatorAddress:   evi.VoteB.ValidatorAddress,
				ValidatorIndex:     evi.VoteB.ValidatorIndex,
				Signature:          evi.VoteB.Signature,
				Extension:          evi.VoteB.Extension,
				ExtensionSignature: evi.VoteB.ExtensionSignature,
			},
			TotalVotingPower: evi.TotalVotingPower,
			ValidatorPower:   evi.ValidatorPower,
			Timestamp:        timestamppb.New(evi.Timestamp),
		}

		return &tmtypes.Evidence{
			Sum: &tmtypes.Evidence_DuplicateVoteEvidence{
				DuplicateVoteEvidence: pbev,
			},
		}, nil

	case *cmttypes.LightClientAttackEvidence:
		pbev := &tmtypes.LightClientAttackEvidence{
			ConflictingBlock: &tmtypes.LightBlock{
				SignedHeader: &tmtypes.SignedHeader{
					Header: headerToProto(evi.ConflictingBlock.SignedHeader.Header),
				},
				ValidatorSet: &tmtypes.ValidatorSet{},
			},
			CommonHeight:        evi.CommonHeight,
			ByzantineValidators: []*tmtypes.Validator{},
			TotalVotingPower:    evi.TotalVotingPower,
			Timestamp:           timestamppb.New(evi.Timestamp),
		}

		for _, val := range evi.ConflictingBlock.ValidatorSet.Validators {
			pkey, err := pubKeyToProto(val.PubKey)
			if err != nil {
				return nil, err
			}

			pbev.ConflictingBlock.ValidatorSet.Validators = append(pbev.ConflictingBlock.ValidatorSet.Validators, &tmtypes.Validator{
				Address:     val.Address,
				PubKey:      pkey,
				VotingPower: val.VotingPower,
			})
		}

		for _, val := range evi.ByzantineValidators {
			pkey, err := pubKeyToProto(val.PubKey)
			if err != nil {
				return nil, err
			}

			pbev.ByzantineValidators = append(pbev.ByzantineValidators, &tmtypes.Validator{
				Address:     val.Address,
				PubKey:      pkey,
				VotingPower: val.VotingPower,
			})
		}

		return &tmtypes.Evidence{
			Sum: &tmtypes.Evidence_LightClientAttackEvidence{
				LightClientAttackEvidence: pbev,
			},
		}, nil

	default:
		return nil, fmt.Errorf("toproto: evidence is not recognized: %T", evi)
	}
}

// pubKeyToProto takes crypto.PubKey and transforms it to a protobuf Pubkey
func pubKeyToProto(k crypto.PubKey) (*tmcrypto.PublicKey, error) {
	var kp *tmcrypto.PublicKey
	switch k := k.(type) {
	case ed25519.PubKey:
		kp = &tmcrypto.PublicKey{
			Sum: &tmcrypto.PublicKey_Ed25519{
				Ed25519: k,
			},
		}
	case secp256k1.PubKey:
		kp = &tmcrypto.PublicKey{
			Sum: &tmcrypto.PublicKey_Secp256K1{
				Secp256K1: k,
			},
		}
	default:
		return kp, fmt.Errorf("toproto: key type %v is not supported", k)
	}
	return kp, nil
}

func blockToProto(block *cmttypes.Block) *tmtypes.Block {
	b := &tmtypes.Block{
		Header: &tmtypes.Header{
			Version: &tmversion.Consensus{
				Block: block.Version.Block,
				App:   block.Version.App,
			},
			ChainId: block.ChainID,
			Height:  block.Height,
			Time:    timestamppb.New(block.Time),
			LastBlockId: &tmtypes.BlockID{
				Hash: block.LastBlockID.Hash,
				PartSetHeader: &tmtypes.PartSetHeader{
					Total: block.LastBlockID.PartSetHeader.Total,
					Hash:  block.LastBlockID.PartSetHeader.Hash,
				},
			},
			LastCommitHash:     block.LastCommitHash,
			DataHash:           block.DataHash,
			ValidatorsHash:     block.ValidatorsHash,
			NextValidatorsHash: block.NextValidatorsHash,
			ConsensusHash:      block.ConsensusHash,
			AppHash:            block.AppHash,
			LastResultsHash:    block.LastResultsHash,
			EvidenceHash:       block.EvidenceHash,
			ProposerAddress:    block.ProposerAddress,
		},
		Data: &tmtypes.Data{
			Txs: [][]byte{},
		},
		Evidence: &tmtypes.EvidenceList{},
		LastCommit: &tmtypes.Commit{
			Height: block.LastCommit.Height,
			Round:  block.LastCommit.Round,
			BlockId: &tmtypes.BlockID{
				Hash: block.LastCommit.BlockID.Hash,
				PartSetHeader: &tmtypes.PartSetHeader{
					Total: block.LastCommit.BlockID.PartSetHeader.Total,
					Hash:  block.LastCommit.BlockID.PartSetHeader.Hash,
				},
			},
			Signatures: []*tmtypes.CommitSig{},
		},
	}

	for _, tx := range block.Data.Txs {
		b.Data.Txs = append(b.Data.Txs, tx)
	}

	for _, sig := range block.LastCommit.Signatures {
		b.LastCommit.Signatures = append(b.LastCommit.Signatures, &tmtypes.CommitSig{
			BlockIdFlag:      tmtypes.BlockIDFlag(sig.BlockIDFlag),
			ValidatorAddress: sig.ValidatorAddress,
			Timestamp:        timestamppb.New(sig.Timestamp),
			Signature:        sig.Signature,
		})
	}

	for _, ev := range block.Evidence.Evidence {
		pbev, err := evidenceToProto(ev)
		if err != nil {
			panic(err)
		}

		b.Evidence.Evidence = append(b.Evidence.Evidence, pbev)
	}

	return b
}

// convertHeader converts CometBFT header to sdk header
func headerToSdkHeader(h cmttypes.Header, consAddrCdc address.Codec) *tmv1beta1.Header {
	propAddress, err := consAddrCdc.BytesToString(h.ProposerAddress)
	if err != nil {
		panic(err)
	}

	return &tmv1beta1.Header{
		Version: &tmversion.Consensus{
			Block: h.Version.Block,
			App:   h.Version.App,
		},
		ChainId: h.ChainID,
		Height:  h.Height,
		Time:    timestamppb.New(h.Time),
		LastBlockId: &tmtypes.BlockID{
			Hash: h.LastBlockID.Hash,
			PartSetHeader: &tmtypes.PartSetHeader{
				Total: h.LastBlockID.PartSetHeader.Total,
				Hash:  h.LastBlockID.PartSetHeader.Hash,
			},
		},
		LastCommitHash:     h.LastCommitHash,
		DataHash:           h.DataHash,
		ValidatorsHash:     h.ValidatorsHash,
		NextValidatorsHash: h.NextValidatorsHash,
		ConsensusHash:      h.ConsensusHash,
		AppHash:            h.AppHash,
		LastResultsHash:    h.LastResultsHash,
		EvidenceHash:       h.EvidenceHash,
		ProposerAddress:    propAddress,
	}
}

// convertBlock converts CometBFT block to sdk block
func blockToSdkBlock(cmtblock *cmttypes.Block, consAddrCdc address.Codec) *tmv1beta1.Block {
	b := new(tmv1beta1.Block)

	b.Header = headerToSdkHeader(cmtblock.Header, consAddrCdc)
	b.LastCommit = &tmtypes.Commit{
		Height: cmtblock.LastCommit.Height,
		Round:  cmtblock.LastCommit.Round,
		BlockId: &tmtypes.BlockID{
			Hash: cmtblock.LastCommit.BlockID.Hash,
			PartSetHeader: &tmtypes.PartSetHeader{
				Total: cmtblock.LastCommit.BlockID.PartSetHeader.Total,
				Hash:  cmtblock.LastCommit.BlockID.PartSetHeader.Hash,
			},
		},
		Signatures: []*tmtypes.CommitSig{},
	}

	for _, sig := range cmtblock.LastCommit.Signatures {
		b.LastCommit.Signatures = append(b.LastCommit.Signatures, &tmtypes.CommitSig{
			BlockIdFlag:      tmtypes.BlockIDFlag(sig.BlockIDFlag),
			ValidatorAddress: sig.ValidatorAddress,
			Timestamp:        timestamppb.New(sig.Timestamp),
			Signature:        sig.Signature,
		})
	}

	b.Data = &tmtypes.Data{
		Txs: [][]byte{},
	}

	for _, tx := range cmtblock.Data.Txs {
		b.Data.Txs = append(b.Data.Txs, tx)
	}

	b.Evidence = &tmtypes.EvidenceList{
		Evidence: []*tmtypes.Evidence{},
	}

	for _, ev := range cmtblock.Evidence.Evidence {
		pbev, err := evidenceToProto(ev)
		if err != nil {
			panic(err)
		}

		b.Evidence.Evidence = append(b.Evidence.Evidence, pbev)
	}

	return b
}

// ToABCIRequestQuery converts a gRPC ABCIQueryRequest type to an ABCI
// RequestQuery type.
func ToABCIRequestQuery(req *tmv1beta1.ABCIQueryRequest) *abci.RequestQuery {
	return &abci.RequestQuery{
		Data:   req.Data,
		Path:   req.Path,
		Height: req.Height,
		Prove:  req.Prove,
	}
}

// FromABCIResponseQuery converts an ABCI ResponseQuery type to a gRPC
// ABCIQueryResponse type.
func FromABCIResponseQuery(res *abci.ResponseQuery) *tmv1beta1.ABCIQueryResponse {
	var proofOps *tmv1beta1.ProofOps

	if res.ProofOps != nil {
		proofOps = &tmv1beta1.ProofOps{
			Ops: make([]*tmv1beta1.ProofOp, len(res.ProofOps.Ops)),
		}
		for i, proofOp := range res.ProofOps.Ops {
			proofOps.Ops[i] = &tmv1beta1.ProofOp{
				Type_: proofOp.Type,
				Key:   proofOp.Key,
				Data:  proofOp.Data,
			}
		}
	}

	return &tmv1beta1.ABCIQueryResponse{
		Code:      res.Code,
		Log:       res.Log,
		Info:      res.Info,
		Index:     res.Index,
		Key:       res.Key,
		Value:     res.Value,
		ProofOps:  proofOps,
		Height:    res.Height,
		Codespace: res.Codespace,
	}
}
