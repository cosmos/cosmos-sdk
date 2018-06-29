package hd

import (
	"encoding/hex"
	"fmt"
	"github.com/cosmos/cosmos-sdk/crypto/keys/bip39"
)

//nolint
func ExampleStringifyPathParams() {
	path := NewParams(44, 0, 0, false, 0)
	fmt.Println(path.String())
	// Output: 44'/0'/0'/0/0
}

//nolint
func ExampleSomeBIP32TestVecs() {

	seed := bip39.MnemonicToSeed("barrel original fuel morning among eternal " +
		"filter ball stove pluck matrix mechanic")
	master, ch := ComputeMastersFromSeed(seed)
	fmt.Println("keys from fundraiser test-vector (cosmos, bitcoin, ether)")
	fmt.Println()
	// cosmos
	priv, _ := DerivePrivateKeyForPath(master, ch, FullFundraiserPath)
	fmt.Println(hex.EncodeToString(priv[:]))
	// bitcoin
	priv, _ = DerivePrivateKeyForPath(master, ch, "44'/0'/0'/0/0")
	fmt.Println(hex.EncodeToString(priv[:]))
	// ether
	priv, _ = DerivePrivateKeyForPath(master, ch, "44'/60'/0'/0/0")
	fmt.Println(hex.EncodeToString(priv[:]))

	fmt.Println()
	fmt.Println("keys generated via https://coinomi.com/recovery-phrase-tool.html")
	fmt.Println()

	seed = bip39.MnemonicToSeed(
		"advice process birth april short trust crater change bacon monkey medal garment " +
			"gorilla ranch hour rival razor call lunar mention taste vacant woman sister")
	master, ch = ComputeMastersFromSeed(seed)
	priv, _ = DerivePrivateKeyForPath(master, ch, "44'/1'/1'/0/4")
	fmt.Println(hex.EncodeToString(priv[:]))

	seed = bip39.MnemonicToSeed("idea naive region square margin day captain habit " +
		"gun second farm pact pulse someone armed")
	master, ch = ComputeMastersFromSeed(seed)
	priv, _ = DerivePrivateKeyForPath(master, ch, "44'/0'/0'/0/420")
	fmt.Println(hex.EncodeToString(priv[:]))

	fmt.Println()
	fmt.Println("BIP 32 example")
	fmt.Println()

	// bip32 path: m/0/7
	seed = bip39.MnemonicToSeed("monitor flock loyal sick object grunt duty ride develop assault harsh history")
	master, ch = ComputeMastersFromSeed(seed)
	priv, _ = DerivePrivateKeyForPath(master, ch, "0/7")
	fmt.Println(hex.EncodeToString(priv[:]))

	// Output: keys from fundraiser test-vector (cosmos, bitcoin, ether)
	//
	// bfcb217c058d8bbafd5e186eae936106ca3e943889b0b4a093ae13822fd3170c
	// e77c3de76965ad89997451de97b95bb65ede23a6bf185a55d80363d92ee37c3d
	// 7fc4d8a8146dea344ba04c593517d3f377fa6cded36cd55aee0a0bb968e651bc
	//
	// keys generated via https://coinomi.com/recovery-phrase-tool.html
	//
	// a61f10c5fecf40c084c94fa54273b6f5d7989386be4a37669e6d6f7b0169c163
	// 32c4599843de3ef161a629a461d12c60b009b676c35050be5f7ded3a3b23501f
	//
	// BIP 32 example
	//
	// c4c11d8c03625515905d7e89d25dfc66126fbc629ecca6db489a1a72fc4bda78
}
