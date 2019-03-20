package hd

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/cosmos/go-bip39"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var defaultBIP39Passphrase = ""

// return bip39 seed with empty passphrase
func mnemonicToSeed(mnemonic string) []byte {
	return bip39.NewSeed(mnemonic, defaultBIP39Passphrase)
}

//nolint
func ExampleStringifyPathParams() {
	path := NewParams(44, 0, 0, false, 0)
	fmt.Println(path.String())
	path = NewParams(44, 33, 7, true, 9)
	fmt.Println(path.String())
	// Output:
	// 44'/0'/0'/0/0
	// 44'/33'/7'/1/9
}

func TestStringifyFundraiserPathParams(t *testing.T) {
	path := NewFundraiserParams(4, 22)
	require.Equal(t, "44'/118'/4'/0/22", path.String())

	path = NewFundraiserParams(4, 57)
	require.Equal(t, "44'/118'/4'/0/57", path.String())
}

func TestPathToArray(t *testing.T) {
	path := NewParams(44, 118, 1, false, 4)
	require.Equal(t, "[44 118 1 0 4]", fmt.Sprintf("%v", path.DerivationPath()))

	path = NewParams(44, 118, 2, true, 15)
	require.Equal(t, "[44 118 2 1 15]", fmt.Sprintf("%v", path.DerivationPath()))
}

func TestParamsFromPath(t *testing.T) {
	goodCases := []struct {
		params *BIP44Params
		path   string
	}{
		{&BIP44Params{44, 0, 0, false, 0}, "44'/0'/0'/0/0"},
		{&BIP44Params{44, 1, 0, false, 0}, "44'/1'/0'/0/0"},
		{&BIP44Params{44, 0, 1, false, 0}, "44'/0'/1'/0/0"},
		{&BIP44Params{44, 0, 0, true, 0}, "44'/0'/0'/1/0"},
		{&BIP44Params{44, 0, 0, false, 1}, "44'/0'/0'/0/1"},
		{&BIP44Params{44, 1, 1, true, 1}, "44'/1'/1'/1/1"},
		{&BIP44Params{44, 118, 52, true, 41}, "44'/118'/52'/1/41"},
	}

	for i, c := range goodCases {
		params, err := NewParamsFromPath(c.path)
		errStr := fmt.Sprintf("%d %v", i, c)
		assert.NoError(t, err, errStr)
		assert.EqualValues(t, c.params, params, errStr)
		assert.Equal(t, c.path, c.params.String())
	}

	badCases := []struct {
		path string
	}{
		{"43'/0'/0'/0/0"},   // doesnt start with 44
		{"44'/1'/0'/0/0/5"}, // too many fields
		{"44'/0'/1'/0"},     // too few fields
		{"44'/0'/0'/2/0"},   // change field can only be 0/1
		{"44/0'/0'/0/0"},    // first field needs '
		{"44'/0/0'/0/0"},    // second field needs '
		{"44'/0'/0/0/0"},    // third field needs '
		{"44'/0'/0'/0'/0"},  // fourth field must not have '
		{"44'/0'/0'/0/0'"},  // fifth field must not have '
		{"44'/-1'/0'/0/0"},  // no negatives
		{"44'/0'/0'/-1/0"},  // no negatives
		{"a'/0'/0'/-1/0"},   // valid values
		{"0/X/0'/-1/0"},     // valid values
		{"44'/0'/X/-1/0"},   // valid values
		{"44'/0'/0'/%/0"},   // valid values
		{"44'/0'/0'/0/%"},   // valid values
	}

	for i, c := range badCases {
		params, err := NewParamsFromPath(c.path)
		errStr := fmt.Sprintf("%d %v", i, c)
		assert.Nil(t, params, errStr)
		assert.Error(t, err, errStr)
	}

}

//nolint
func ExampleSomeBIP32TestVecs() {

	seed := mnemonicToSeed("barrel original fuel morning among eternal " +
		"filter ball stove pluck matrix mechanic")
	master, ch := ComputeMastersFromSeed(seed)
	fmt.Println("keys from fundraiser test-vector (cosmos, bitcoin, ether)")
	fmt.Println()
	// cosmos
	priv, err := DerivePrivateKeyForPath(master, ch, FullFundraiserPath)
	if err != nil {
		fmt.Println("INVALID")
	} else {
		fmt.Println(hex.EncodeToString(priv[:]))
	}
	// bitcoin
	priv, err = DerivePrivateKeyForPath(master, ch, "44'/0'/0'/0/0")
	if err != nil {
		fmt.Println("INVALID")
	} else {
		fmt.Println(hex.EncodeToString(priv[:]))
	}
	// ether
	priv, err = DerivePrivateKeyForPath(master, ch, "44'/60'/0'/0/0")
	if err != nil {
		fmt.Println("INVALID")
	} else {
		fmt.Println(hex.EncodeToString(priv[:]))
	}
	// INVALID
	priv, err = DerivePrivateKeyForPath(master, ch, "X/0'/0'/0/0")
	if err != nil {
		fmt.Println("INVALID")
	} else {
		fmt.Println(hex.EncodeToString(priv[:]))
	}
	priv, err = DerivePrivateKeyForPath(master, ch, "-44/0'/0'/0/0")
	if err != nil {
		fmt.Println("INVALID")
	} else {
		fmt.Println(hex.EncodeToString(priv[:]))
	}

	fmt.Println()
	fmt.Println("keys generated via https://coinomi.com/recovery-phrase-tool.html")
	fmt.Println()

	seed = mnemonicToSeed(
		"advice process birth april short trust crater change bacon monkey medal garment " +
			"gorilla ranch hour rival razor call lunar mention taste vacant woman sister")
	master, ch = ComputeMastersFromSeed(seed)
	priv, _ = DerivePrivateKeyForPath(master, ch, "44'/1'/1'/0/4")
	fmt.Println(hex.EncodeToString(priv[:]))

	seed = mnemonicToSeed("idea naive region square margin day captain habit " +
		"gun second farm pact pulse someone armed")
	master, ch = ComputeMastersFromSeed(seed)
	priv, _ = DerivePrivateKeyForPath(master, ch, "44'/0'/0'/0/420")
	fmt.Println(hex.EncodeToString(priv[:]))

	fmt.Println()
	fmt.Println("BIP 32 example")
	fmt.Println()

	// bip32 path: m/0/7
	seed = mnemonicToSeed("monitor flock loyal sick object grunt duty ride develop assault harsh history")
	master, ch = ComputeMastersFromSeed(seed)
	priv, _ = DerivePrivateKeyForPath(master, ch, "0/7")
	fmt.Println(hex.EncodeToString(priv[:]))

	// Output: keys from fundraiser test-vector (cosmos, bitcoin, ether)
	//
	// bfcb217c058d8bbafd5e186eae936106ca3e943889b0b4a093ae13822fd3170c
	// e77c3de76965ad89997451de97b95bb65ede23a6bf185a55d80363d92ee37c3d
	// 7fc4d8a8146dea344ba04c593517d3f377fa6cded36cd55aee0a0bb968e651bc
	// INVALID
	// INVALID
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
