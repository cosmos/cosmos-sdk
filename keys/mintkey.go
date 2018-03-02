package keys

import (
	"encoding/hex"
	"fmt"

	cmn "github.com/tendermint/tmlibs/common"

	"github.com/tendermint/go-crypto"
	"github.com/tendermint/go-crypto/keys/bcrypt"
)

const (
	blockTypePrivKey = "TENDERMINT PRIVATE KEY"
	blockTypeKeyInfo = "TENDERMINT KEY INFO"
)

func armorInfoBytes(bz []byte) string {
	header := map[string]string{
		"type":    "Info",
		"version": "0.0.0",
	}
	armorStr := crypto.EncodeArmor(blockTypeKeyInfo, header, bz)
	return armorStr
}

func unarmorInfoBytes(armorStr string) (bz []byte, err error) {
	blockType, header, bz, err := crypto.DecodeArmor(armorStr)
	if err != nil {
		return
	}
	if blockType != blockTypeKeyInfo {
		err = fmt.Errorf("Unrecognized armor type: %v", blockType)
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
	key, err := bcrypt.GenerateFromPassword(saltBytes, []byte(passphrase), 12) // TODO parameterize.  12 is good today (2016)
	if err != nil {
		cmn.Exit("Error generating bcrypt key from passphrase: " + err.Error())
	}
	key = crypto.Sha256(key) // Get 32 bytes
	privKeyBytes := privKey.Bytes()
	return saltBytes, crypto.EncryptSymmetric(privKeyBytes, key)
}

func decryptPrivKey(saltBytes []byte, encBytes []byte, passphrase string) (privKey crypto.PrivKey, err error) {
	key, err := bcrypt.GenerateFromPassword(saltBytes, []byte(passphrase), 12) // TODO parameterize.  12 is good today (2016)
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
