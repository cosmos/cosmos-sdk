package keeper

type AddressType uint8

const (
	// AddressTypeCosmos represents a native Cosmos address, derived using the Cosmos secp256k1 derivation if it represents a secp256k1 key.
	AddressTypeCosmos AddressType = iota
	// AddressTypeCosmosEVM represents a native Cosmos address, derived using the EVM secp256k1 derivation and representing a secp256k1 key.
	// An account will never have this address type if it does not have a secp256k1 key.
	AddressTypeCosmosEVM
	// AddressTypeEVM represents an EVM address, derived using the EVM secp256k1 derivation if it represents a secp256k1 key.
	AddressTypeEVM
)
