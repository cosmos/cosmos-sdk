package baseapp

import (
	"time"

	abci "github.com/cometbft/cometbft/abci/types"

	"cosmossdk.io/core/comet"
)

type reqBeginBlockInfo struct {
	*abci.RequestBeginBlock
}

func (r reqBeginBlockInfo) Evidence() []comet.Misbehavior {
	return misbehaviorWrapperList(r.ByzantineValidators)
}

func misbehaviorWrapperList(validators []abci.Misbehavior) []comet.Misbehavior {
	misbehaviors := make([]comet.Misbehavior, len(validators))
	for i, v := range validators {
		misbehaviors[i] = misbehaviorWrapper{v}
	}
	return misbehaviors
}

func (r reqBeginBlockInfo) ValidatorsHash() []byte {
	return r.Header.ValidatorsHash
}

func (r reqBeginBlockInfo) ProposerAddress() []byte {
	return r.Header.ProposerAddress
}

func (r reqBeginBlockInfo) DecidedLastCommit() comet.CommitInfo {
	return commitInfoWrapper{r.LastCommitInfo}
}

var _ comet.BlockInfo = (*reqBeginBlockInfo)(nil)

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
	return validateWrapper{v.VoteInfo.Validator}
}

type validateWrapper struct {
	abci.Validator
}

func (v validateWrapper) Address() []byte {
	return v.Validator.Address
}

func (v validateWrapper) Power() int64 {
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
	return validateWrapper{m.Misbehavior.Validator}
}

func (m misbehaviorWrapper) Time() time.Time {
	return m.Misbehavior.Time
}

func (m misbehaviorWrapper) TotalVotingPower() int64 {
	return m.Misbehavior.TotalVotingPower
}

type reqPrepareProposalInfo struct {
	*abci.RequestPrepareProposal
}

func (r reqPrepareProposalInfo) Evidence() []comet.Misbehavior {
	//TODO implement me
	panic("implement me")
}

func (r reqPrepareProposalInfo) ValidatorsHash() []byte {
	//TODO implement me
	panic("implement me")
}

func (r reqPrepareProposalInfo) ProposerAddress() []byte {
	return r.RequestPrepareProposal.ProposerAddress
}

func (r reqPrepareProposalInfo) DecidedLastCommit() comet.CommitInfo {
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
	return validateWrapper{e.ExtendedVoteInfo.Validator}
}

type reqProcessProposalInfo struct {
	*abci.RequestProcessProposal
}

func (r reqProcessProposalInfo) Evidence() []comet.Misbehavior {
	return misbehaviorWrapperList(r.Misbehavior)
}

func (r reqProcessProposalInfo) ValidatorsHash() []byte {
	// TODO: is this correct???
	return r.RequestProcessProposal.NextValidatorsHash
}

func (r reqProcessProposalInfo) ProposerAddress() []byte {
	return r.RequestProcessProposal.ProposerAddress
}

func (r reqProcessProposalInfo) DecidedLastCommit() comet.CommitInfo {
	// TODO: is this correct???
	return commitInfoWrapper{r.RequestProcessProposal.ProposedLastCommit}
}

var _ comet.BlockInfo = &reqProcessProposalInfo{}
