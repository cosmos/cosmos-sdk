package exported

import (
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
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
	GetPrefix() commitmenttypes.MerklePrefix
	ValidateBasic() error
}
