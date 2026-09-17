package crypto_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
	"testing"

	cmtcrypto "github.com/cometbft/cometbft/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/core/address"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log/v2"

	"github.com/cosmos/cosmos-sdk/codec"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	"github.com/cosmos/cosmos-sdk/crypto"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/bcrypt"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/crypto/xsalsa20symmetric"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	"github.com/cosmos/cosmos-sdk/types"
)

func TestArmorUnarmorPrivKey(t *testing.T) {
	priv := secp256k1.GenPrivKey()
	armored := crypto.EncryptArmorPrivKey(priv, "passphrase", "")
	_, _, err := crypto.UnarmorDecryptPrivKey(armored, "wrongpassphrase")
	require.Error(t, err)
	decrypted, algo, err := crypto.UnarmorDecryptPrivKey(armored, "passphrase")
	require.NoError(t, err)
	require.Equal(t, string(hd.Secp256k1Type), algo)
	require.True(t, priv.Equals(decrypted))

	// empty string
	decrypted, algo, err = crypto.UnarmorDecryptPrivKey("", "passphrase")
	require.Error(t, err)
	require.True(t, errors.Is(err, io.EOF))
	require.Nil(t, decrypted)
	require.Empty(t, algo)

	// wrong key type
	armored = crypto.ArmorPubKeyBytes(priv.PubKey().Bytes(), "")
	_, _, err = crypto.UnarmorDecryptPrivKey(armored, "passphrase")
	require.Error(t, err)
	require.Contains(t, err.Error(), "unrecognized armor type")

	// armor key manually
	encryptPrivKeyFn := func(privKey cryptotypes.PrivKey, passphrase string) (saltBytes, encBytes []byte) {
		saltBytes = cmtcrypto.CRandBytes(16)
		key, err := bcrypt.GenerateFromPassword(saltBytes, []byte(passphrase), crypto.BcryptSecurityParameter)
		require.NoError(t, err)
		key = cmtcrypto.Sha256(key) // get 32 bytes
		privKeyBytes := legacy.Cdc.Amino.MustMarshalBinaryBare(privKey)
		return saltBytes, xsalsa20symmetric.EncryptSymmetric(privKeyBytes, key)
	}
	saltBytes, encBytes := encryptPrivKeyFn(priv, "passphrase")

	// wrong kdf header
	headerWrongKdf := map[string]string{
		"kdf":  "wrong",
		"salt": fmt.Sprintf("%X", saltBytes),
		"type": "secp256k",
	}
	armored = crypto.EncodeArmor("TENDERMINT PRIVATE KEY", headerWrongKdf, encBytes)
	_, _, err = crypto.UnarmorDecryptPrivKey(armored, "passphrase")
	require.Error(t, err)
	require.Equal(t, "unrecognized KDF type: wrong", err.Error())
}

func TestArmorUnarmorPubKey(t *testing.T) {
	// Select the encryption and storage for your cryptostore
	var cdc codec.Codec

	err := depinject.Inject(depinject.Configs(
		configurator.NewAppConfig(),
		depinject.Supply(log.NewNopLogger(),
			func() address.Codec { return addresscodec.NewBech32Codec("cosmos") },
			func() runtime.ValidatorAddressCodec { return addresscodec.NewBech32Codec("cosmosvaloper") },
			func() runtime.ConsensusAddressCodec { return addresscodec.NewBech32Codec("cosmosvalcons") },
		),
	), &cdc)
	require.NoError(t, err)

	cstore := keyring.NewInMemory(cdc)

	// Add keys and see they return in alphabetical order
	k, _, err := cstore.NewMnemonic("Bob", keyring.English, types.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	require.NoError(t, err)
	key, err := k.GetPubKey()
	require.NoError(t, err)
	armored := crypto.ArmorPubKeyBytes(legacy.Cdc.Amino.MustMarshalBinaryBare(key), "")
	pubBytes, algo, err := crypto.UnarmorPubKeyBytes(armored)
	require.NoError(t, err)
	pub, err := legacy.PubKeyFromBytes(pubBytes)
	require.NoError(t, err)
	require.Equal(t, string(hd.Secp256k1Type), algo)
	require.True(t, pub.Equals(key))

	armored = crypto.ArmorPubKeyBytes(legacy.Cdc.Amino.MustMarshalBinaryBare(key), "unknown")
	pubBytes, algo, err = crypto.UnarmorPubKeyBytes(armored)
	require.NoError(t, err)
	pub, err = legacy.PubKeyFromBytes(pubBytes)
	require.NoError(t, err)
	require.Equal(t, "unknown", algo)
	require.True(t, pub.Equals(key))

	armored, err = cstore.ExportPrivKeyArmor("Bob", "passphrase")
	require.NoError(t, err)
	_, _, err = crypto.UnarmorPubKeyBytes(armored)
	require.Error(t, err)
	require.Equal(t, `couldn't unarmor bytes: unrecognized armor type "TENDERMINT PRIVATE KEY", expected: "TENDERMINT PUBLIC KEY"`, err.Error())

	// armor pubkey manually
	header := map[string]string{
		"version": "0.0.0",
		"type":    "unknown",
	}
	armored = crypto.EncodeArmor("TENDERMINT PUBLIC KEY", header, pubBytes)
	_, algo, err = crypto.UnarmorPubKeyBytes(armored)
	require.NoError(t, err)
	// return secp256k1 if version is 0.0.0
	require.Equal(t, "secp256k1", algo)

	// missing version header
	header = map[string]string{
		"type": "unknown",
	}
	armored = crypto.EncodeArmor("TENDERMINT PUBLIC KEY", header, pubBytes)
	bz, algo, err := crypto.UnarmorPubKeyBytes(armored)
	require.Nil(t, bz)
	require.Empty(t, algo)
	require.Error(t, err)
	require.Equal(t, "header's version field is empty", err.Error())

	// unknown version header
	header = map[string]string{
		"type":    "unknown",
		"version": "unknown",
	}
	armored = crypto.EncodeArmor("TENDERMINT PUBLIC KEY", header, pubBytes)
	bz, algo, err = crypto.UnarmorPubKeyBytes(armored)
	require.Nil(t, bz)
	require.Empty(t, algo)
	require.Error(t, err)
	require.Equal(t, "unrecognized version: unknown", err.Error())
}

func TestArmorInfoBytes(t *testing.T) {
	bs := []byte("test")
	armoredString := crypto.ArmorInfoBytes(bs)
	unarmoredBytes, err := crypto.UnarmorInfoBytes(armoredString)
	require.NoError(t, err)
	require.True(t, bytes.Equal(bs, unarmoredBytes))
}

func TestUnarmorInfoBytesErrors(t *testing.T) {
	unarmoredBytes, err := crypto.UnarmorInfoBytes("")
	require.Error(t, err)
	require.True(t, errors.Is(err, io.EOF))
	require.Nil(t, unarmoredBytes)

	header := map[string]string{
		"type":    "Info",
		"version": "0.0.1",
	}
	unarmoredBytes, err = crypto.UnarmorInfoBytes(crypto.EncodeArmor(
		"TENDERMINT KEY INFO", header, []byte("plain-text")))
	require.Error(t, err)
	require.Equal(t, "unrecognized version: 0.0.1", err.Error())
	require.Nil(t, unarmoredBytes)
}

func BenchmarkBcryptGenerateFromPassword(b *testing.B) {
	passphrase := []byte("passphrase")
	for securityParam := uint32(9); securityParam < 16; securityParam++ {
		param := securityParam
		b.Run(fmt.Sprintf("benchmark-security-param-%d", param), func(b *testing.B) {
			b.ReportAllocs()
			saltBytes := cmtcrypto.CRandBytes(16)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := bcrypt.GenerateFromPassword(saltBytes, passphrase, param)
				require.Nil(b, err)
			}
		})
	}
}

func TestArmor(t *testing.T) {
	blockType := "MINT TEST"
	data := []byte("somedata")
	armorStr := crypto.EncodeArmor(blockType, nil, data)

	// Decode armorStr and test for equivalence.
	blockType2, _, data2, err := crypto.DecodeArmor(armorStr)
	require.Nil(t, err, "%+v", err)
	assert.Equal(t, blockType, blockType2)
	assert.Equal(t, data, data2)
}

// TestArmorChecksum checks that a block whose contents disagree with its
// CRC-24 footer is rejected. The armor decoder does not check the footer
// itself, and a public key has no authentication tag to fail later, so a
// flipped byte in the key material would otherwise be imported silently as a
// different key.
func TestArmorChecksum(t *testing.T) {
	algo := string(hd.Secp256k1Type)
	pubBytes := legacy.Cdc.MustMarshal(secp256k1.GenPrivKey().PubKey())

	// Flip a bit in the last byte of the key material. The result is still a
	// well-formed public key of the right length, so only the footer can
	// reveal that it is not the key that was exported.
	corrupt := slices.Clone(pubBytes)
	corrupt[len(corrupt)-1] ^= 1

	split := func(s string) []string { return strings.Split(strings.TrimRight(s, "\n"), "\n") }
	join := func(lines []string) string { return strings.Join(lines, "\n") + "\n" }
	withFooter := func(lines []string, footer string) string {
		out := slices.Clone(lines)
		out[len(out)-2] = footer
		return join(out)
	}

	good := split(crypto.ArmorPubKeyBytes(pubBytes, algo))
	bad := split(crypto.ArmorPubKeyBytes(corrupt, algo))
	require.True(t, strings.HasPrefix(good[len(good)-2], "="), "expected a CRC-24 footer")

	// Either block decodes on its own, which is what makes the footer
	// necessary in the first place.
	for _, lines := range [][]string{good, bad} {
		bz, gotAlgo, err := crypto.UnarmorPubKeyBytes(join(lines))
		require.NoError(t, err)
		require.Equal(t, algo, gotAlgo)
		require.Len(t, bz, len(pubBytes))
	}

	// The corrupted body carrying the original footer, which is what a damaged
	// or tampered-with export looks like.
	_, _, err := crypto.UnarmorPubKeyBytes(withFooter(bad, good[len(good)-2]))
	require.ErrorIs(t, err, crypto.ErrArmorChecksum)

	// Footers that cannot match: unreadable, or another block's. The previous
	// decoder skipped its checksum check outright for "=E3J=", which let a
	// tampered block suppress the check by rewriting the footer.
	for _, footer := range []string{"=!!!!", "=E3J=", "=AAAA"} {
		_, _, err := crypto.UnarmorPubKeyBytes(withFooter(good, footer))
		require.ErrorIs(t, err, crypto.ErrArmorChecksum, footer)
	}

	// A block with no footer stays acceptable: RFC 9580 deprecated it, so
	// third-party tooling may legitimately omit it.
	_, gotAlgo, err := crypto.UnarmorPubKeyBytes(join(slices.Delete(slices.Clone(good), len(good)-2, len(good)-1)))
	require.NoError(t, err)
	require.Equal(t, algo, gotAlgo)

	// A footer belonging to a different block in the same input must not fail
	// a block that is intact.
	_, _, got, err := crypto.DecodeArmor(join(good) + crypto.EncodeArmor("MINT TEST", nil, []byte("other")))
	require.NoError(t, err)
	require.Equal(t, pubBytes, got)

	// Nor may a header the decoder parses loosely, such as one with no space
	// after its colon, cause the check to be skipped.
	loose := split(strings.Replace(join(good), "version: 0.0.1", "version:00.0.1", 1))
	require.Contains(t, join(loose), "version:00.0.1")
	_, _, got, err = crypto.DecodeArmor(join(loose))
	require.NoError(t, err)
	require.Equal(t, pubBytes, got)
	_, _, _, err = crypto.DecodeArmor(withFooter(loose, "=AAAA"))
	require.ErrorIs(t, err, crypto.ErrArmorChecksum)
}

func TestBcryptLegacyEncryption(t *testing.T) {
	privKey := secp256k1.GenPrivKey()
	saltBytes := cmtcrypto.CRandBytes(16)
	passphrase := "passphrase"
	privKeyBytes := legacy.Cdc.MustMarshal(privKey)

	// Bcrypt + Aead
	headerBcrypt := map[string]string{
		"kdf":  "bcrypt",
		"salt": fmt.Sprintf("%X", saltBytes),
	}
	keyBcrypt, _ := bcrypt.GenerateFromPassword(saltBytes, []byte(passphrase), 12) // Legacy key generation
	keyBcrypt = cmtcrypto.Sha256(keyBcrypt)

	// bcrypt + xsalsa20symmetric
	encBytesBcryptXsalsa20symmetric := xsalsa20symmetric.EncryptSymmetric(privKeyBytes, keyBcrypt)

	type testCase struct {
		description string
		armor       string
	}

	for _, scenario := range []testCase{
		{
			description: "Argon2 + Aead",
			armor:       crypto.EncryptArmorPrivKey(privKey, "passphrase", ""),
		},
		{
			description: "Bcrypt + xsalsa20symmetric",
			armor:       crypto.EncodeArmor("TENDERMINT PRIVATE KEY", headerBcrypt, encBytesBcryptXsalsa20symmetric),
		},
	} {
		t.Run(scenario.description, func(t *testing.T) {
			_, _, err := crypto.UnarmorDecryptPrivKey(scenario.armor, "wrongpassphrase")
			require.Error(t, err)
			decryptedPrivKey, _, err := crypto.UnarmorDecryptPrivKey(scenario.armor, "passphrase")
			require.NoError(t, err)
			require.True(t, privKey.Equals(decryptedPrivKey))
		})
	}

	// Test wrong kdf header
	headerWithoutKdf := map[string]string{
		"kdf":  "wrongKdf",
		"salt": fmt.Sprintf("%X", saltBytes),
	}

	_, _, err := crypto.UnarmorDecryptPrivKey(crypto.EncodeArmor("TENDERMINT PRIVATE KEY", headerWithoutKdf, encBytesBcryptXsalsa20symmetric), "passphrase")
	require.Error(t, err)
	require.Equal(t, "unrecognized KDF type: wrongKdf", err.Error())
}
