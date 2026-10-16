package crypto_test

import (
	"bytes"
	"errors"
	"fmt"
	"hash/adler32"
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

// errArmor is the error a malformed armored block is rejected with.
const errArmor = "openpgp: invalid data: armor invalid"

// TestArmorChecksum checks that a block whose contents disagree with its
// CRC-24 footer is rejected. A public key has no authentication tag to fail
// later, so a flipped byte in the key material would otherwise be imported
// silently as a different key.
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
	footer := good[len(good)-2]
	require.True(t, strings.HasPrefix(footer, "="), "expected a CRC-24 footer")

	// Either block decodes on its own, which is what makes the footer
	// necessary in the first place.
	for _, lines := range [][]string{good, bad} {
		bz, gotAlgo, err := crypto.UnarmorPubKeyBytes(join(lines))
		require.NoError(t, err)
		require.Equal(t, algo, gotAlgo)
		require.Len(t, bz, len(pubBytes))
	}

	// The corrupted body carrying the original footer, which is what a damaged
	// export looks like.
	tampered := withFooter(bad, footer)
	_, _, err := crypto.UnarmorPubKeyBytes(tampered)
	require.EqualError(t, err, "couldn't unarmor bytes: "+errArmor)
	_, _, err = crypto.UnarmorPubKeyBytes(withFooter(good, "=AAAA"))
	require.EqualError(t, err, "couldn't unarmor bytes: "+errArmor)

	// Footers that cannot be read are rejected rather than skipped: one that is
	// not base64, and one padded with whitespace, as copy-pasting through chat
	// or email often does.
	for _, f := range []string{"=!!!!", footer + " ", " " + footer} {
		_, _, err := crypto.UnarmorPubKeyBytes(withFooter(bad, f))
		require.Error(t, err, "%q", f)
	}

	// A block with no footer stays acceptable: RFC 9580 deprecated it, so
	// third-party tooling may legitimately omit it.
	footerless := join(slices.Delete(slices.Clone(good), len(good)-2, len(good)-1))
	_, gotAlgo, err := crypto.UnarmorPubKeyBytes(footerless)
	require.NoError(t, err)
	require.Equal(t, algo, gotAlgo)

	// Whichever block the decoder returns from input holding several, its own
	// footer is the one checked. Each prefix ends in a block the decoder
	// skips, ahead of the tampered one.
	for name, prefix := range map[string]string{
		"block with no type":       "-----BEGIN -----\n=AAAA\n-----END -----\n",
		"headers that don't parse": good[0] + "\nnocolon\n" + strings.Replace(good[0], "-----BEGIN ", "-----END ", 1) + "\n",
		// The decoder reads the nested BEGIN line as a header that does not
		// parse, and resumes after it.
		"nested BEGIN line": good[0] + "\n" + footerless,
		// The decoder skips a line longer than its 100-byte buffer.
		"overlong BEGIN line": good[0] + strings.Repeat(" ", 120) + "\n" + strings.Join(split(footerless)[1:], "\n") + "\n",
	} {
		_, _, _, err := crypto.DecodeArmor(prefix + tampered)
		require.EqualError(t, err, errArmor, name)
	}
}

// The armor tests below cover the code copied from
// golang.org/x/crypto/openpgp/armor into armor.go. TestDecodeEncodeArmor,
// TestDecodeArmorLongHeader and their test vectors are upstream's, covered by
// the license reproduced in armor.go; the rest pin upstream's checksum
// handling, which must not change.

func TestDecodeEncodeArmor(t *testing.T) {
	blockType, headers, data, err := crypto.DecodeArmor(armorExample1)
	require.NoError(t, err)
	require.Equal(t, "PGP SIGNATURE", blockType)
	require.Equal(t, map[string]string{"Version": "GnuPG v1.4.10 (GNU/Linux)"}, headers)
	require.Equal(t, uint32(0x27b144be), adler32.Checksum(data))

	require.Equal(t, armorExample1, crypto.EncodeArmor(blockType, headers, data))
}

func TestDecodeArmorLongHeader(t *testing.T) {
	_, headers, _, err := crypto.DecodeArmor(armorLongLine)
	require.NoError(t, err)
	require.Equal(t, longValueExpected, headers["Version"])
}

// TestDecodeArmorChecksum checks that the body is checked against the checksum
// line, and that a block without one is still accepted, as RFC 9580 deprecated
// it.
func TestDecodeArmorChecksum(t *testing.T) {
	const footer = "=/teI"
	require.Contains(t, armorExample1, "\n"+footer+"\n")

	for _, in := range []string{
		armorExample1,
		strings.Replace(armorExample1, footer+"\n", "", 1),
	} {
		_, _, _, err := crypto.DecodeArmor(in)
		require.NoError(t, err)
	}

	for name, in := range map[string]string{
		// Same body with the last data byte changed, keeping the original
		// checksum line.
		"corrupted body": strings.Replace(armorExample1, "b1g=", "b1k=", 1),
		"other checksum": strings.Replace(armorExample1, footer, "=AAAA", 1),
	} {
		_, _, _, err := crypto.DecodeArmor(in)
		require.EqualError(t, err, errArmor, name)
	}
}

// TestDecodeArmorMalformedChecksumLine checks how a checksum line that cannot
// be read is handled.
func TestDecodeArmorMalformedChecksumLine(t *testing.T) {
	for _, line := range []string{
		// Not base64.
		"=!!!!",
		// Padded with whitespace, as copy-pasting through chat or email
		// often does: read as body, which is not valid base64.
		"=/teI ",
		"=/teI\t",
	} {
		_, _, _, err := crypto.DecodeArmor(strings.Replace(armorExample1, "=/teI", line, 1))
		require.Error(t, err, "%q", line)
	}

	// A line that decodes to fewer than three bytes is not treated as a
	// checksum line, so the block is accepted unchecked. This is upstream's
	// behavior, kept as is.
	_, _, _, err := crypto.DecodeArmor(strings.Replace(armorExample1, "=/teI", "=E3J=", 1))
	require.NoError(t, err)
}

// TestDecodeArmorChecksumLineMustEndBlock checks that the armor trailer has to
// follow the checksum line directly.
func TestDecodeArmorChecksumLineMustEndBlock(t *testing.T) {
	for name, in := range map[string]string{
		"line after checksum": strings.Replace(armorExample1, "=/teI\n", "=/teI\nAAAA\n", 1),
		"no trailer":          strings.TrimSuffix(armorExample1, "-----END PGP SIGNATURE-----"),
	} {
		_, _, _, err := crypto.DecodeArmor(in)
		require.EqualError(t, err, errArmor, name)
	}
}

const armorExample1 = `-----BEGIN PGP SIGNATURE-----
Version: GnuPG v1.4.10 (GNU/Linux)

iJwEAAECAAYFAk1Fv/0ACgkQo01+GMIMMbsYTwQAiAw+QAaNfY6WBdplZ/uMAccm
4g+81QPmTSGHnetSb6WBiY13kVzK4HQiZH8JSkmmroMLuGeJwsRTEL4wbjRyUKEt
p1xwUZDECs234F1xiG5enc5SGlRtP7foLBz9lOsjx+LEcA4sTl5/2eZR9zyFZqWW
TxRjs+fJCIFuo71xb1g=
=/teI
-----END PGP SIGNATURE-----`

const armorLongLine = `-----BEGIN PGP SIGNATURE-----
Version: 0123456789abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz

iQEcBAABAgAGBQJMtFESAAoJEKsQXJGvOPsVj40H/1WW6jaMXv4BW+1ueDSMDwM8
kx1fLOXbVM5/Kn5LStZNt1jWWnpxdz7eq3uiqeCQjmqUoRde3YbB2EMnnwRbAhpp
cacnAvy9ZQ78OTxUdNW1mhX5bS6q1MTEJnl+DcyigD70HG/yNNQD7sOPMdYQw0TA
byQBwmLwmTsuZsrYqB68QyLHI+DUugn+kX6Hd2WDB62DKa2suoIUIHQQCd/ofwB3
WfCYInXQKKOSxu2YOg2Eb4kLNhSMc1i9uKUWAH+sdgJh7NBgdoE4MaNtBFkHXRvv
okWuf3+xA9ksp1npSY/mDvgHijmjvtpRDe6iUeqfCn8N9u9CBg8geANgaG8+QA4=
=wfQG
-----END PGP SIGNATURE-----`

const longValueExpected = "0123456789abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz"

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
