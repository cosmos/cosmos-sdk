//nolint
package gov

import "github.com/cosmos/cosmos-sdk/x/gov/types"

const (
	StatusNil           = types.StatusNil
	StatusDepositPeriod = types.StatusDepositPeriod
	StatusVotingPeriod  = types.StatusVotingPeriod
	StatusPassed        = types.StatusPassed
	StatusRejected      = types.StatusRejected

	DefaultCodespace  = types.DefaultCodespace
	DefaultParamspace = types.DefaultParamspace
	RouterKey         = types.RouterKey
	StoreKey          = types.StoreKey
	QuerierRoute      = types.QuerierRoute
)

type (
	Content  = types.Content
	Handler  = types.Handler
	Proposal = types.Proposal

	Deposit    = types.Deposit
	Vote       = types.Vote
	VoteOption = types.VoteOption

	ProposalStatus = types.ProposalStatus

	MsgSubmitProposal = types.MsgSubmitProposal
	MsgDeposit        = types.MsgDeposit
	MsgVote           = types.MsgVote

	TallyResult = types.TallyResult
)

var (
	ErrUnknownProposal         = types.ErrUnknownProposal
	ErrInactiveProposal        = types.ErrInactiveProposal
	ErrAlreadyActiveProposal   = types.ErrAlreadyActiveProposal
	ErrAlreadyFinishedProposal = types.ErrAlreadyFinishedProposal
	ErrAddressNotStaked        = types.ErrAddressNotStaked
	ErrInvalidTitle            = types.ErrInvalidTitle
	ErrInvalidDescription      = types.ErrInvalidDescription
	ErrInvalidProposalType     = types.ErrInvalidProposalType
	ErrInvalidVote             = types.ErrInvalidVote
	ErrInvalidGenesis          = types.ErrInvalidGenesis
	ErrNoProposalHandlerExists = types.ErrNoProposalHandlerExists

	NewProposal     = types.NewProposal
	ProposalHandler = types.ProposalHandler

	ValidVoteOption     = types.ValidVoteOption
	ValidProposalStatus = types.ValidProposalStatus

	EmptyTallyResult = types.EmptyTallyResult

	RegisterCodec = types.RegisterCodec
)
