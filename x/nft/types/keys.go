package types

const (
	// module name
	ModuleName = "nft"

	// StoreKey is the default store key for nft
	StoreKey = ModuleName
)

var (
	TypeKey = []byte{0x01}
	NFTKey  = []byte{0x02}
)

// GetTypeKey returns the byte representation of the nft type key
func GetTypeKey(typ string) []byte {
	return append(TypeKey, []byte(typ)...)
}

// GetNFTKey returns the byte representation of the nft
func GetNFTKey(typ string) []byte {
	return append(NFTKey, []byte(typ)...)
}

func GetNFTIdKey(id string) []byte {
	return []byte(id)
}
