package keys

import (
	"encoding/hex"
	"fmt"

	cmn "github.com/tendermint/tendermint/libs/common"

	"github.com/cosmos/cosmos-sdk/crypto/keys/bcrypt"
	"github.com/tendermint/tendermint/crypto"
)

const (
	blockTypePrivKey = "TENDERMINT PRIVATE KEY"
	blockTypeKeyInfo = "TENDERMINT KEY INFO"
	blockTypePubKey  = "TENDERMINT PUBLIC KEY"
)

// Make bcrypt security parameter var, so it can be changed within the lcd test
// Making the bcrypt security parameter a var shouldn't be a security issue:
// One can't verify an invalid key by maliciously changing the bcrypt
// parameter during a runtime vulnerability. The main security
// threat this then exposes would be something that changes this during
// runtime before the user creates their key. This vulnerability must
// succeed to update this to that same value before every subsequent call
// to gaiacli keys in future startups / or the attacker must get access
// to the filesystem. However, with a similar threat model (changing
// variables in runtime), one can cause the user to sign a different tx
// than what they see, which is a significantly cheaper attack then breaking
// a bcrypt hash. (Recall that the nonce still exists to break rainbow tables)
// TODO: Consider increasing default
var BcryptSecurityParameter = 12

func armorInfoBytes(bz []byte) string {
	return armorBytes(bz, blockTypeKeyInfo)
}

func armorPubKeyBytes(bz []byte) string {
	return armorBytes(bz, blockTypePubKey)
}

func armorBytes(bz []byte, blockType string) string {
	header := map[string]string{
		"type":    "Info",
		"version": "0.0.0",
	}
	return crypto.EncodeArmor(blockType, header, bz)
}

func unarmorInfoBytes(armorStr string) (bz []byte, err error) {
	return unarmorBytes(armorStr, blockTypeKeyInfo)
}

func unarmorPubKeyBytes(armorStr string) (bz []byte, err error) {
	return unarmorBytes(armorStr, blockTypePubKey)
}

func unarmorBytes(armorStr, blockType string) (bz []byte, err error) {
	bType, header, bz, err := crypto.DecodeArmor(armorStr)
	if err != nil {
		return
	}
	if bType != blockType {
		err = fmt.Errorf("Unrecognized armor type %q, expected: %q", bType, blockType)
		return
	}
	if header["version"] != "0.0.0" {
		err = fmt.Errorf("Unrecognized version: %v", header["version"])
		return
	}
	return
}

func encryptArmorPrivKey(privKey crypto.PrivKey, passphrase string) string {
	saltBytes, encBytes := encryptPrivKey(privKey, passphrase)
	header := map[string]string{
		"kdf":  "bcrypt",
		"salt": fmt.Sprintf("%X", saltBytes),
	}
	armorStr := crypto.EncodeArmor(blockTypePrivKey, header, encBytes)
	return armorStr
}

func unarmorDecryptPrivKey(armorStr string, passphrase string) (crypto.PrivKey, error) {
	var privKey crypto.PrivKey
	blockType, header, encBytes, err := crypto.DecodeArmor(armorStr)
	if err != nil {
		return privKey, err
	}
	if blockType != blockTypePrivKey {
		return privKey, fmt.Errorf("Unrecognized armor type: %v", blockType)
	}
	if header["kdf"] != "bcrypt" {
		return privKey, fmt.Errorf("Unrecognized KDF type: %v", header["KDF"])
	}
	if header["salt"] == "" {
		return privKey, fmt.Errorf("Missing salt bytes")
	}
	saltBytes, err := hex.DecodeString(header["salt"])
	if err != nil {
		return privKey, fmt.Errorf("Error decoding salt: %v", err.Error())
	}
	privKey, err = decryptPrivKey(saltBytes, encBytes, passphrase)
	return privKey, err
}

func encryptPrivKey(privKey crypto.PrivKey, passphrase string) (saltBytes []byte, encBytes []byte) {
	saltBytes = crypto.CRandBytes(16)
	key, err := bcrypt.GenerateFromPassword(saltBytes, []byte(passphrase), BcryptSecurityParameter)
	if err != nil {
		cmn.Exit("Error generating bcrypt key from passphrase: " + err.Error())
	}
	key = crypto.Sha256(key) // Get 32 bytes
	privKeyBytes := privKey.Bytes()
	return saltBytes, crypto.EncryptSymmetric(privKeyBytes, key)
}

func decryptPrivKey(saltBytes []byte, encBytes []byte, passphrase string) (privKey crypto.PrivKey, err error) {
	key, err := bcrypt.GenerateFromPassword(saltBytes, []byte(passphrase), BcryptSecurityParameter)
	if err != nil {
		cmn.Exit("Error generating bcrypt key from passphrase: " + err.Error())
	}
	key = crypto.Sha256(key) // Get 32 bytes
	privKeyBytes, err := crypto.DecryptSymmetric(encBytes, key)
	if err != nil {
		return privKey, err
	}
	privKey, err = crypto.PrivKeyFromBytes(privKeyBytes)
	return privKey, err
}
