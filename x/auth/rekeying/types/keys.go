package types

var (
	// module name
	ModuleName = "rekeying"

	// StoreKey is string representation of the store key for changepubkey
	StoreKey = ModuleName

	// QuerierRoute is the querier route for changepubkey
	QuerierRoute = ModuleName

	// route key
	RouterKey = ModuleName

	// KeyPrefixPubKeyHistory defines history of PubKey history of an account
	KeyPrefixPubKeyHistory = []byte{0x01} // prefix for the timestamps in pubkey history queue
)
