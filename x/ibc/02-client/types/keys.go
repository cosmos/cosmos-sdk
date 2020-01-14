package types

const (
	// SubModuleName defines the IBC client name
	SubModuleName string = "client"

	// StoreKey is the store key string for IBC client
	StoreKey string = SubModuleName

	// RouterKey is the message route for IBC client
	RouterKey string = SubModuleName

	// QuerierRoute is the querier route for IBC client
	QuerierRoute string = SubModuleName
)
