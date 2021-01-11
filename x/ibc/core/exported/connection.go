package exported

// ConnectionI describes the required methods for a connection.
type ConnectionI interface {
	GetClientID() string
	GetState() int32
	GetCounterparty() CounterpartyConnectionI
	GetVersions() []Version
	GetDelayPeriod() uint64
	ValidateBasic() error
}

// CounterpartyConnectionI describes the required methods for a counterparty connection.
type CounterpartyConnectionI interface {
	GetClientID() string
	GetConnectionID() string
	GetPrefix() Prefix
	ValidateBasic() error
}

// Version defines an IBC version used in connection handshake negotiation.
type Version interface {
	GetIdentifier() string
	GetFeatures() []string
	VerifyProposedVersion(Version) error
}
