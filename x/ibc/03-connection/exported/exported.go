package exported

// ConnectionI describes the required methods for a connection.
type ConnectionI interface {
	GetState() StateI
	GetClientID() string
	GetCounterparty() CounterpartyI
	GetVersions() []string
	ValidateBasic() error
}

// CounterpartyI describes the required methods for a counterparty connection.
type CounterpartyI interface {
	GetClientID() string
	GetConnectionID() string
	ValidateBasic() error
}

// StateI describes the required methods for a connection state.
type StateI interface {
	String() string
	MarshalJSON() ([]byte, error)
	UnmarshalJSON(data []byte) error
}