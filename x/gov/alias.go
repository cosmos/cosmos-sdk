package gov

// nolint

import (
	"github.com/cosmos/cosmos-sdk/x/gov/keeper"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	MaxDescriptionLength  = types.MaxDescriptionLength
	MaxTitleLength        = types.MaxTitleLength
	DefaultPeriod         = types.DefaultPeriod
	ModuleName            = types.ModuleName
	StoreKey              = types.StoreKey
	RouterKey             = types.RouterKey
	QuerierRoute          = types.QuerierRoute
	DefaultParamspace     = types.DefaultParamspace
	TypeMsgDeposit        = types.TypeMsgDeposit
	TypeMsgVote           = types.TypeMsgVote
	TypeMsgSubmitProposal = types.TypeMsgSubmitProposal
	StatusNil             = types.StatusNil
	StatusDepositPeriod   = types.StatusDepositPeriod
	StatusVotingPeriod    = types.StatusVotingPeriod
	StatusPassed          = types.StatusPassed
	StatusRejected        = types.StatusRejected
	StatusFailed          = types.StatusFailed
	ProposalTypeText      = types.ProposalTypeText
	QueryParams           = types.QueryParams
	QueryProposals        = types.QueryProposals
	QueryProposal         = types.QueryProposal
	QueryDeposits         = types.QueryDeposits
	QueryDeposit          = types.QueryDeposit
	QueryVotes            = types.QueryVotes
	QueryVote             = types.QueryVote
	QueryTally            = types.QueryTally
	ParamDeposit          = types.ParamDeposit
	ParamVoting           = types.ParamVoting
	ParamTallying         = types.ParamTallying
	OptionEmpty           = types.OptionEmpty
	OptionYes             = types.OptionYes
	OptionAbstain         = types.OptionAbstain
	OptionNo              = types.OptionNo
	OptionNoWithVeto      = types.OptionNoWithVeto
)

var (
	// functions aliases
	RegisterInvariants            = keeper.RegisterInvariants
	AllInvariants                 = keeper.AllInvariants
	ModuleAccountInvariant        = keeper.ModuleAccountInvariant
	NewKeeper                     = keeper.NewKeeper
	NewQuerier                    = keeper.NewQuerier
	RegisterCodec                 = types.RegisterCodec
	RegisterProposalTypeCodec     = types.RegisterProposalTypeCodec
	ValidateAbstract              = types.ValidateAbstract
	NewDeposit                    = types.NewDeposit
	ErrUnknownProposal            = types.ErrUnknownProposal
	ErrInactiveProposal           = types.ErrInactiveProposal
	ErrAlreadyActiveProposal      = types.ErrAlreadyActiveProposal
	ErrInvalidProposalContent     = types.ErrInvalidProposalContent
	ErrInvalidProposalType        = types.ErrInvalidProposalType
	ErrInvalidVote                = types.ErrInvalidVote
	ErrInvalidGenesis             = types.ErrInvalidGenesis
	ErrNoProposalHandlerExists    = types.ErrNoProposalHandlerExists
	NewGenesisState               = types.NewGenesisState
	DefaultGenesisState           = types.DefaultGenesisState
	ValidateGenesis               = types.ValidateGenesis
	GetProposalIDBytes            = types.GetProposalIDBytes
	GetProposalIDFromBytes        = types.GetProposalIDFromBytes
	ProposalKey                   = types.ProposalKey
	ActiveProposalByTimeKey       = types.ActiveProposalByTimeKey
	ActiveProposalQueueKey        = types.ActiveProposalQueueKey
	InactiveProposalByTimeKey     = types.InactiveProposalByTimeKey
	InactiveProposalQueueKey      = types.InactiveProposalQueueKey
	DepositsKey                   = types.DepositsKey
	DepositKey                    = types.DepositKey
	VotesKey                      = types.VotesKey
	VoteKey                       = types.VoteKey
	SplitProposalKey              = types.SplitProposalKey
	SplitActiveProposalQueueKey   = types.SplitActiveProposalQueueKey
	SplitInactiveProposalQueueKey = types.SplitInactiveProposalQueueKey
	SplitKeyDeposit               = types.SplitKeyDeposit
	SplitKeyVote                  = types.SplitKeyVote
	NewMsgSubmitProposal          = types.NewMsgSubmitProposal
	NewMsgDeposit                 = types.NewMsgDeposit
	NewMsgVote                    = types.NewMsgVote
	ParamKeyTable                 = types.ParamKeyTable
	NewDepositParams              = types.NewDepositParams
	NewTallyParams                = types.NewTallyParams
	NewVotingParams               = types.NewVotingParams
	NewParams                     = types.NewParams
	NewProposal                   = types.NewProposal
	NewRouter                     = types.NewRouter
	ProposalStatusFromString      = types.ProposalStatusFromString
	ValidProposalStatus           = types.ValidProposalStatus
	NewTextProposal               = types.NewTextProposal
	RegisterProposalType          = types.RegisterProposalType
	ContentFromProposalType       = types.ContentFromProposalType
	IsValidProposalType           = types.IsValidProposalType
	ProposalHandler               = types.ProposalHandler
	NewQueryProposalParams        = types.NewQueryProposalParams
	NewQueryDepositParams         = types.NewQueryDepositParams
	NewQueryVoteParams            = types.NewQueryVoteParams
	NewQueryProposalsParams       = types.NewQueryProposalsParams
	NewValidatorGovInfo           = types.NewValidatorGovInfo
	NewTallyResult                = types.NewTallyResult
	NewTallyResultFromMap         = types.NewTallyResultFromMap
	EmptyTallyResult              = types.EmptyTallyResult
	NewVote                       = types.NewVote
	VoteOptionFromString          = types.VoteOptionFromString
	ValidVoteOption               = types.ValidVoteOption

	// variable aliases
	ModuleCdc                   = types.ModuleCdc
	ProposalsKeyPrefix          = types.ProposalsKeyPrefix
	ActiveProposalQueuePrefix   = types.ActiveProposalQueuePrefix
	InactiveProposalQueuePrefix = types.InactiveProposalQueuePrefix
	ProposalIDKey               = types.ProposalIDKey
	DepositsKeyPrefix           = types.DepositsKeyPrefix
	VotesKeyPrefix              = types.VotesKeyPrefix
	ParamStoreKeyDepositParams  = types.ParamStoreKeyDepositParams
	ParamStoreKeyVotingParams   = types.ParamStoreKeyVotingParams
	ParamStoreKeyTallyParams    = types.ParamStoreKeyTallyParams
)

type (
	Keeper               = keeper.Keeper
	Content              = types.Content
	Handler              = types.Handler
	Deposit              = types.Deposit
	Deposits             = types.Deposits
	GenesisState         = types.GenesisState
	MsgSubmitProposal    = types.MsgSubmitProposal
	MsgDeposit           = types.MsgDeposit
	MsgVote              = types.MsgVote
	DepositParams        = types.DepositParams
	TallyParams          = types.TallyParams
	VotingParams         = types.VotingParams
	Params               = types.Params
	Proposal             = types.Proposal
	Proposals            = types.Proposals
	ProposalQueue        = types.ProposalQueue
	ProposalStatus       = types.ProposalStatus
	TextProposal         = types.TextProposal
	QueryProposalParams  = types.QueryProposalParams
	QueryDepositParams   = types.QueryDepositParams
	QueryVoteParams      = types.QueryVoteParams
	QueryProposalsParams = types.QueryProposalsParams
	ValidatorGovInfo     = types.ValidatorGovInfo
	TallyResult          = types.TallyResult
	Vote                 = types.Vote
	Votes                = types.Votes
	VoteOption           = types.VoteOption
)
