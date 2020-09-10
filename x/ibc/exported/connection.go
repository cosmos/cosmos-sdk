package exported

// ConnectionI describes the required methods for a connection.
type ConnectionI interface {
	GetClientID() string
	GetState() int32
	GetCounterparty() CounterpartyConnectionI
	GetVersions() []string
	ValidateBasic() error
}

// CounterpartyConnectionI describes the required methods for a counterparty connection.
type CounterpartyConnectionI interface {
	GetClientID() string
	GetConnectionID() string
	GetPrefix() Prefix
	ValidateBasic() error
}
