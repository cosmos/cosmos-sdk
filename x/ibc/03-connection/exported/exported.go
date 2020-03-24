package exported

import (
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	commitmentexported "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
)

// ConnectionI describes the required methods for a connection.
type ConnectionI interface {
	GetState() types.State
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
