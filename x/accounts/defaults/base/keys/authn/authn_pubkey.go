package authn

import (
	"bytes"
	"crypto"
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"

	cometcrypto "github.com/cometbft/cometbft/crypto"
	"github.com/cosmos/gogoproto/proto"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

type Signature struct {
	AuthenticatorData string `json:"authenticatorData"`
	ClientDataJSON    string `json:"clientDataJSON"`
	Signature         string `json:"signature"`
}

const (
	keyType    = "authn"
	PubKeyName = "tendermint/PubKeyAuthn"
)

var (
	_ cryptotypes.PubKey = (*AuthnPubKey)(nil)
)

const AuthnPubKeySize = 33

func (pubKey *AuthnPubKey) Address() cometcrypto.Address {
	if len(pubKey.Key) != AuthnPubKeySize {
		panic("length of pubkey is incorrect")
	}

	return address.Hash(proto.MessageName(pubKey), pubKey.Key)
}

func (pubKey *AuthnPubKey) Bytes() []byte {
	return pubKey.Key
}

func (pubKey *AuthnPubKey) String() string {
	return fmt.Sprintf("PubKeyAuthn{%X}", pubKey.Key)
}

func (pubKey *AuthnPubKey) Type() string {
	return keyType
}

func (pubKey *AuthnPubKey) Equals(other cryptotypes.PubKey) bool {
	return pubKey.Type() == other.Type() && bytes.Equal(pubKey.Bytes(), other.Bytes())
}

func (pubKey *AuthnPubKey) VerifySignature(msg, sigStr []byte) bool {
	sig := Signature{}
	err := json.Unmarshal(sigStr, &sig)
	if err != nil {
		return false
	}

	clientDataJSON, err := hex.DecodeString(sig.ClientDataJSON)
	if err != nil {
		return false
	}

	clientData := make(map[string]interface{})
	err = json.Unmarshal(clientDataJSON, &clientData)
	if err != nil {
		return false
	}

	challengeBase64, ok := clientData["challenge"].(string)
	if !ok {
		return false
	}
	challenge, err := base64.RawURLEncoding.DecodeString(challengeBase64)
	if err != nil {
		return false
	}

	// Check challenge == msg
	if !bytes.Equal(challenge, msg) {
		return false
	}

	publicKey := &ecdsa.PublicKey{Curve: elliptic.P256()}

	publicKey.X, publicKey.Y = elliptic.UnmarshalCompressed(elliptic.P256(), pubKey.Key)
	if publicKey.X == nil || publicKey.Y == nil {
		return false
	}

	signatureBytes, err := hex.DecodeString(sig.Signature)
	if err != nil {
		return false
	}

	authenticatorData, err := hex.DecodeString(sig.AuthenticatorData)
	if err != nil {
		return false
	}

	// check authenticatorData length
	if len(authenticatorData) < 37 {
		return false
	}

	clientDataHash := sha256.Sum256(clientDataJSON)
	payload := append(authenticatorData, clientDataHash[:]...)

	h := crypto.SHA256.New()
	h.Write(payload)

	return ecdsa.VerifyASN1(publicKey, h.Sum(nil), signatureBytes)
}
