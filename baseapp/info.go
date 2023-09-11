package baseapp

import (
	abci "github.com/cometbft/cometbft/abci/types"

	"cosmossdk.io/core/comet"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type prepareProposalInfo struct {
	*abci.RequestPrepareProposal
}

var _ comet.BlockInfo = (*prepareProposalInfo)(nil)

func (r prepareProposalInfo) GetEvidence() comet.EvidenceList {
	return sdk.EvidenceWrapper{Evidence: r.Misbehavior}
}

func (r prepareProposalInfo) GetValidatorsHash() []byte {
	return r.NextValidatorsHash
}

func (r prepareProposalInfo) GetProposerAddress() []byte {
	return r.RequestPrepareProposal.ProposerAddress
}

func (r prepareProposalInfo) GetLastCommit() comet.CommitInfo {
	return extendedCommitInfoWrapper{r.RequestPrepareProposal.LocalLastCommit}
}

var _ comet.BlockInfo = (*prepareProposalInfo)(nil)

type extendedCommitInfoWrapper struct {
	abci.ExtendedCommitInfo
}

var _ comet.CommitInfo = (*extendedCommitInfoWrapper)(nil)

func (e extendedCommitInfoWrapper) Round() int32 {
	return e.ExtendedCommitInfo.Round
}

func (e extendedCommitInfoWrapper) Votes() comet.VoteInfos {
	return extendedVoteInfoWrapperList{e.ExtendedCommitInfo.Votes}
}

type extendedVoteInfoWrapperList struct {
	votes []abci.ExtendedVoteInfo
}

var _ comet.VoteInfos = (*extendedVoteInfoWrapperList)(nil)

func (e extendedVoteInfoWrapperList) Len() int {
	return len(e.votes)
}

func (e extendedVoteInfoWrapperList) Get(i int) comet.VoteInfo {
	return extendedVoteInfoWrapper{e.votes[i]}
}

type extendedVoteInfoWrapper struct {
	abci.ExtendedVoteInfo
}

var _ comet.VoteInfo = (*extendedVoteInfoWrapper)(nil)

func (e extendedVoteInfoWrapper) GetBlockIDFlag() comet.BlockIDFlag {
	return comet.BlockIDFlag(e.ExtendedVoteInfo.BlockIdFlag)
}

func (e extendedVoteInfoWrapper) Validator() comet.Validator {
	return sdk.ValidatorWrapper{Validator: e.ExtendedVoteInfo.Validator}
}
