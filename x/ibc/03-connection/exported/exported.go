package exported

import (
	commitmentexported "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
)

// ConnectionI describes the required methods for a connection.
type ConnectionI interface {
	GetID() string
	GetClientID() string
	GetState() int32
	GetCounterparty() CounterpartyI
	GetVersions() []string
	ValidateBasic() error
}

// CounterpartyI describes the required methods for a counterparty connection.
type CounterpartyI interface {
	GetClientID() string
	GetConnectionID() string
	GetPrefix() (commitmentexported.Prefix, error)
	ValidateBasic() error
}
