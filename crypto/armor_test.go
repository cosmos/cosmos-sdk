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

// TestArmorChecksum checks that a body which disagrees with its CRC-24 footer
// is rejected. The armor decoder does not verify the footer itself, and a
// public key has no authentication tag to fail later, so a flipped byte in the
// key material would otherwise be imported silently as a different key.
func TestArmorChecksum(t *testing.T) {
	algo := string(hd.Secp256k1Type)
	pubBytes := legacy.Cdc.MustMarshal(secp256k1.GenPrivKey().PubKey())

	// Flip a bit in the last byte of the key material. The result is still a
	// well-formed public key of the correct length, so only the footer can
	// reveal that it is not the key that was exported.
	corruptedBytes := slices.Clone(pubBytes)
	corruptedBytes[len(corruptedBytes)-1] ^= 0x01

	good := crypto.ArmorPubKeyBytes(pubBytes, algo)
	corrupted := crypto.ArmorPubKeyBytes(corruptedBytes, algo)

	// Both decode on their own, which is what makes the footer necessary.
	for name, armorStr := range map[string]string{"original": good, "corrupted": corrupted} {
		bz, gotAlgo, err := crypto.UnarmorPubKeyBytes(armorStr)
		require.NoError(t, err, name)
		require.Equal(t, algo, gotAlgo, name)
		require.Len(t, bz, len(pubBytes), name)
	}

	// Splice the original footer onto the corrupted body, which is what a
	// damaged or tampered-with export looks like.
	lines := strings.Split(strings.TrimRight(corrupted, "\n"), "\n")
	goodLines := strings.Split(strings.TrimRight(good, "\n"), "\n")
	footer := len(lines) - 2
	require.True(t, strings.HasPrefix(lines[footer], "="), "expected a CRC-24 footer, got %q", lines[footer])
	lines[footer] = goodLines[len(goodLines)-2]

	_, _, err := crypto.UnarmorPubKeyBytes(strings.Join(lines, "\n") + "\n")
	require.ErrorIs(t, err, crypto.ErrArmorChecksum)

	// A block with no footer at all stays acceptable: RFC 9580 deprecated it,
	// so third-party tooling may legitimately omit it.
	_, gotAlgo, err := crypto.UnarmorPubKeyBytes(strings.Join(slices.Delete(goodLines, footer, footer+1), "\n") + "\n")
	require.NoError(t, err)
	require.Equal(t, algo, gotAlgo)
}

// TestArmorChecksumLooseHeaders locks in that the footer is still verified for
// blocks whose headers the decoder accepts loosely. The decoder splits headers
// on a bare colon, so the footer scan has to agree with it about which lines
// are headers: if it were stricter it would go looking for the footer of a
// block the decoder never read, and skip the check.
func TestArmorChecksumLooseHeaders(t *testing.T) {
	const blockType = "MINT TEST"
	data := []byte("some key material that is long enough to wrap onto a second base64 line")

	// A header with no space after the colon, which the decoder still accepts.
	loose := strings.Replace(
		crypto.EncodeArmor(blockType, map[string]string{"version": "0.0.1"}, data),
		"version: 0.0.1", "version:00.0.1", 1)
	require.Contains(t, loose, "version:00.0.1")

	gotType, _, got, err := crypto.DecodeArmor(loose)
	require.NoError(t, err)
	require.Equal(t, blockType, gotType)
	require.Equal(t, data, got)

	lines := strings.Split(strings.TrimRight(loose, "\n"), "\n")
	body := 3 // BEGIN line, header, blank line, then the body
	require.Len(t, lines[body], 64, "expected a full-width base64 body line")

	// Corrupting the body of that same block must still be caught.
	corrupted := []byte(lines[body])
	corrupted[3] ^= 1
	withCorruptBody := slices.Clone(lines)
	withCorruptBody[body] = string(corrupted)
	_, _, _, err = crypto.DecodeArmor(strings.Join(withCorruptBody, "\n") + "\n")
	require.ErrorIs(t, err, crypto.ErrArmorChecksum)

	// A footer that is present but unreadable is a malformed block, not a
	// block without a footer.
	withBadFooter := slices.Clone(lines)
	withBadFooter[len(withBadFooter)-2] = "=!!!!"
	_, _, _, err = crypto.DecodeArmor(strings.Join(withBadFooter, "\n") + "\n")
	require.ErrorIs(t, err, crypto.ErrArmorCorrupt)

	// So is a footer with no trailer after it.
	_, _, _, err = crypto.DecodeArmor(strings.Join(lines[:len(lines)-1], "\n") + "\n")
	require.ErrorIs(t, err, crypto.ErrArmorCorrupt)

	// A footer that is valid base64 but does not decode to three bytes is
	// malformed too. The previous decoder skipped the checksum check outright
	// for these, which let a tampered block suppress its own integrity check
	// by rewriting the footer.
	withPaddedFooter := slices.Clone(lines)
	withPaddedFooter[len(withPaddedFooter)-2] = "=E3J="
	_, _, _, err = crypto.DecodeArmor(strings.Join(withPaddedFooter, "\n") + "\n")
	require.ErrorIs(t, err, crypto.ErrArmorCorrupt)
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
