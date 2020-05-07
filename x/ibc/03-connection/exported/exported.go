package exported

import (
	commitmentexported "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/common"
)

// ConnectionI describes the required methods for a connection.
type ConnectionI interface {
	GetState() common.State
	GetID() string
	GetClientID() string
	GetCounterparty() CounterpartyI
	GetVersions() []string
	ValidateBasic() error
}

// CounterpartyI describes the required methods for a counterparty connection.
type CounterpartyI interface {
	GetClientID() string
	GetConnectionID() string
	GetPrefix() commitmentexported.Prefix
	ValidateBasic() error
}
