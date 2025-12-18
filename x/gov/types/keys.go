package types

import (
	"cosmossdk.io/collections"
	collcodec "cosmossdk.io/collections/codec"
)

const (
	// ModuleName is the name of the module
	ModuleName = "gov"

	// StoreKey is the store key string for gov
	StoreKey = ModuleName

	// RouterKey is the message route for gov
	RouterKey = ModuleName
)

var (
	ProposalsKeyPrefix            = collections.NewPrefix(0)  // ProposalsKeyPrefix stores the proposals raw bytes.
	ActiveProposalQueuePrefix     = collections.NewPrefix(1)  // ActiveProposalQueuePrefix stores the active proposals.
	InactiveProposalQueuePrefix   = collections.NewPrefix(2)  // InactiveProposalQueuePrefix stores the inactive proposals.
	ProposalIDKey                 = collections.NewPrefix(3)  // ProposalIDKey stores the sequence representing the next proposal ID.
	VotingPeriodProposalKeyPrefix = collections.NewPrefix(4)  // VotingPeriodProposalKeyPrefix stores which proposals are on voting period.
	DepositsKeyPrefix             = collections.NewPrefix(16) // DepositsKeyPrefix stores deposits.
	VotesKeyPrefix                = collections.NewPrefix(32) // VotesKeyPrefix stores the votes of proposals.
	ParamsKey                     = collections.NewPrefix(48) // ParamsKey stores the module's params.

	// NOTE, this is changed compared to normal Cosmos-SDK v0.50 to match AtomOne.
	// When another chain adopts the AtomOne fork, it must set the its current constitution (key 49) to this one.
	ConstitutionKey                          = collections.NewPrefix(64) // ConstitutionKey stores a chain's constitution.
	QuorumCheckQueuePrefix                   = collections.NewPrefix(5)
	LastMinDepositKey                        = collections.NewPrefix(7)
	LastMinInitialDepositKey                 = collections.NewPrefix(9)
	ParticipationEMAKey                      = collections.NewPrefix(80)
	ConstitutionAmendmentParticipationEMAKey = collections.NewPrefix(96)
	LawParticipationEMAKey                   = collections.NewPrefix(112)
	GovernorsKeyPrefix                       = collections.NewPrefix(128)
	GovernanceDelegationKeyPrefix            = collections.NewPrefix(129)
	ValidatorSharesByGovernorKeyPrefix       = collections.NewPrefix(130)
	GovernanceDelegationsByGovernorKeyPrefix = collections.NewPrefix(131)

	// GovernorAddressKey follows the same semantics as AccAddressKey.
	GovernorAddressKey collcodec.KeyCodec[GovernorAddress] = governorAddressKey{
		stringDecoder: GovernorAddressFromBech32,
		keyType:       "gov.GovernorAddress",
	}
)
