package baseapp

import (
	"time"

	abci "github.com/cometbft/cometbft/abci/types"

	"cosmossdk.io/core/comet"
)

type CometInfo struct {
	Misbehavior        []abci.Misbehavior
	NextValidatorsHash []byte
	ProposerAddress    []byte
	LastCommit         abci.CommitInfo
}

func (r CometInfo) GetEvidence() []comet.Misbehavior {
	return misbehaviorWrapperList(r.Misbehavior)
}

func misbehaviorWrapperList(validators []abci.Misbehavior) []comet.Misbehavior {
	misbehaviors := make([]comet.Misbehavior, len(validators))
	for i, v := range validators {
		misbehaviors[i] = misbehaviorWrapper{v}
	}
	return misbehaviors
}

func (r CometInfo) GetValidatorsHash() []byte {
	return r.NextValidatorsHash
}

func (r CometInfo) GetProposerAddress() []byte {
	return r.ProposerAddress
}

func (r CometInfo) GetLastCommit() comet.CommitInfo {
	return commitInfoWrapper{r.LastCommit}
}

var _ comet.BlockInfo = (*CometInfo)(nil)

type commitInfoWrapper struct {
	abci.CommitInfo
}

func (c commitInfoWrapper) Round() int32 {
	return c.CommitInfo.Round
}

func (c commitInfoWrapper) Votes() []comet.VoteInfo {
	return voteInfoWrapperList(c.CommitInfo.Votes)
}

func voteInfoWrapperList(votes []abci.VoteInfo) []comet.VoteInfo {
	voteInfos := make([]comet.VoteInfo, len(votes))
	for i, v := range votes {
		voteInfos[i] = voteInfoWrapper{v}
	}
	return voteInfos
}

type voteInfoWrapper struct {
	abci.VoteInfo
}

func (v voteInfoWrapper) SignedLastBlock() bool {
	return v.VoteInfo.SignedLastBlock
}

func (v voteInfoWrapper) Validator() comet.Validator {
	return validatorWrapper{v.VoteInfo.Validator}
}

type validatorWrapper struct {
	abci.Validator
}

func (v validatorWrapper) Address() []byte {
	return v.Validator.Address
}

func (v validatorWrapper) Power() int64 {
	return v.Validator.Power
}

type misbehaviorWrapper struct {
	abci.Misbehavior
}

func (m misbehaviorWrapper) Type() comet.MisbehaviorType {
	return comet.MisbehaviorType(m.Misbehavior.Type)
}

func (m misbehaviorWrapper) Height() int64 {
	return m.Misbehavior.Height
}

func (m misbehaviorWrapper) Validator() comet.Validator {
	return validatorWrapper{m.Misbehavior.Validator}
}

func (m misbehaviorWrapper) Time() time.Time {
	return m.Misbehavior.Time
}

func (m misbehaviorWrapper) TotalVotingPower() int64 {
	return m.Misbehavior.TotalVotingPower
}

type reqPrepareProposalInfo struct {
	abci.RequestPrepareProposal
}

func (r reqPrepareProposalInfo) GetEvidence() []comet.Misbehavior {
	return misbehaviorWrapperList(r.Misbehavior)

}

func (r reqPrepareProposalInfo) GetValidatorsHash() []byte {
	return r.NextValidatorsHash
}

func (r reqPrepareProposalInfo) GetProposerAddress() []byte {
	return r.RequestPrepareProposal.ProposerAddress
}

func (r reqPrepareProposalInfo) GetLastCommit() comet.CommitInfo {
	return extendedCommitInfoWrapper{r.RequestPrepareProposal.LocalLastCommit}
}

var _ comet.BlockInfo = (*reqPrepareProposalInfo)(nil)

type extendedCommitInfoWrapper struct {
	abci.ExtendedCommitInfo
}

func (e extendedCommitInfoWrapper) Round() int32 {
	return e.ExtendedCommitInfo.Round
}

func (e extendedCommitInfoWrapper) Votes() []comet.VoteInfo {
	return extendedVoteInfoWrapperList(e.ExtendedCommitInfo.Votes)
}

func extendedVoteInfoWrapperList(votes []abci.ExtendedVoteInfo) []comet.VoteInfo {
	voteInfos := make([]comet.VoteInfo, len(votes))
	for i, v := range votes {
		voteInfos[i] = extendedVoteInfoWrapper{v}
	}
	return voteInfos
}

type extendedVoteInfoWrapper struct {
	abci.ExtendedVoteInfo
}

func (e extendedVoteInfoWrapper) SignedLastBlock() bool {
	return e.ExtendedVoteInfo.SignedLastBlock
}

func (e extendedVoteInfoWrapper) Validator() comet.Validator {
	return validatorWrapper{e.ExtendedVoteInfo.Validator}
}
