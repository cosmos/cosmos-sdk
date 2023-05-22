package types

const (
	// ModuleName defines the module name
	ModuleName = "circuit"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName
)

// KVStore keys
var (
	AccountPermissionPrefix = []byte{0x01}
	DisableListPrefix       = []byte{0x02}
)

func CreateAddressPrefix(account []byte) []byte {
	key := make([]byte, len(AccountPermissionPrefix)+len(account)+1)
	copy(key, AccountPermissionPrefix)
	copy(key[len(AccountPermissionPrefix):], account)
	return key
}

func CreateDisableMsgPrefix(msgURL string) []byte {
	key := make([]byte, len(DisableListPrefix)+len(msgURL)+1)
	copy(key, DisableListPrefix)
	copy(key[len(DisableListPrefix):], msgURL)
	return key
}
