package codec

import "os"

// nolint: govet
func ExamplePrintRegisteredTypes() {
	amino.PrintTypes(os.Stdout)
	// Output:
	// | Type | Name | Prefix | Length | Notes |
	// | ---- | ---- | ------ | ----- | ------ |
	// | PubKey | tendermint/PubKeySr25519 | 0x0DFB1005 | variable |  |
	// | PubKey | tendermint/PubKeyEd25519 | 0x1624DE64 | variable |  |
	// | PubKey | tendermint/PubKeySecp256k1 | 0xEB5AE987 | variable |  |
	// | LegacyAminoPubKey | tendermint/PubKeyMultisigThreshold | 0x22C1F7E2 | variable |  |
	// | PrivKey | tendermint/PrivKeySr25519 | 0x2F82D78B | variable |  |
	// | PrivKey | tendermint/PrivKeyEd25519 | 0xA3288910 | variable |  |
	// | PrivKey | tendermint/PrivKeySecp256k1 | 0xE1B0F79B | variable |  |
}
