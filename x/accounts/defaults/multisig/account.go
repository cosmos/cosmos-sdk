package multisig

import (
	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/header"
	"cosmossdk.io/x/accounts/accountstd"
	"cosmossdk.io/x/tx/signing"
)

var (
	ParticipansPrefix  = collections.NewPrefix(0)
	SequencePrefix     = collections.NewPrefix(1)
	ThresholdPrefix    = collections.NewPrefix(2)
	QuorumPrefix       = collections.NewPrefix(3)
	VotingPeriodPrefix = collections.NewPrefix(4)
	RevotePrefix       = collections.NewPrefix(5)
	EarlyExecution     = collections.NewPrefix(6)
)

// Compile-time type assertions
var (
	_ accountstd.Interface = (*Account)(nil)
)

type Account struct {
	Participants   collections.Map[[]byte, uint64]
	Sequence       collections.Sequence
	Threshold      collections.Item[uint64]
	Quorum         collections.Item[uint64]
	VotingPeriod   collections.Item[uint64]
	Revote         collections.Item[bool]
	EarlyExecution collections.Item[bool]

	addrCodec address.Codec
	hs        header.Service

	signingHandlers *signing.HandlerMap
}

func NewAccount(name string, handlerMap *signing.HandlerMap) accountstd.AccountCreatorFunc {
	return func(deps accountstd.Dependencies) (string, accountstd.Interface, error) {
		return name, &Account{
			Participants:    collections.NewMap(deps.SchemaBuilder, ParticipansPrefix, "participants", collections.BytesKey, collections.Uint64Value),
			Sequence:        collections.NewSequence(deps.SchemaBuilder, SequencePrefix, "sequence"),
			Threshold:       collections.NewItem(deps.SchemaBuilder, ThresholdPrefix, "threshold", collections.Uint64Value),
			Quorum:          collections.NewItem(deps.SchemaBuilder, QuorumPrefix, "quorum", collections.Uint64Value),
			VotingPeriod:    collections.NewItem(deps.SchemaBuilder, VotingPeriodPrefix, "voting_period", collections.Uint64Value),
			Revote:          collections.NewItem(deps.SchemaBuilder, RevotePrefix, "revote", collections.BoolValue),
			EarlyExecution:  collections.NewItem(deps.SchemaBuilder, EarlyExecution, "early_execution", collections.BoolValue),
			addrCodec:       deps.AddressCodec,
			signingHandlers: handlerMap,
		}, nil
	}
}

// RegisterExecuteHandlers implements implementation.Account.
func (*Account) RegisterExecuteHandlers(builder *accountstd.ExecuteBuilder) {
	panic("unimplemented")
}

// RegisterInitHandler implements implementation.Account.
func (*Account) RegisterInitHandler(builder *accountstd.InitBuilder) {
	panic("unimplemented")
}

// RegisterQueryHandlers implements implementation.Account.
func (*Account) RegisterQueryHandlers(builder *accountstd.QueryBuilder) {
	panic("unimplemented")
}
