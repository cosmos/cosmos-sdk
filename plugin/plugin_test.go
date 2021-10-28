package plugin_test

import (
	"github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/plugin"
)

type PubKeyModule struct{}

type AuthModule struct{}

type SecpHandler struct{ *secp256k1.PubKey }

func NewSecpHandler(pubKey *secp256k1.PubKey) types.PubKey { return &SecpHandler{PubKey: pubKey} }

type EdHandler struct{ *ed25519.PubKey }

func NewEdHandler(pubKey *ed25519.PubKey) types.PubKey { return &EdHandler{PubKey: pubKey} }

func (m PubKeyModule) Provide(host plugin.Host) {
	host.Register(
		NewSecpHandler,
		NewEdHandler,
	)
}

func (m AuthModule) Provide(host plugin.Host) AuthKeeper {
	host.Expect("cosmos.crypto.PubKey", (*types.PubKey)(nil))
	return AuthKeeper{pluginHost: host}
}

type AuthKeeper struct {
	pluginHost plugin.Host
}

func (k AuthKeeper) GetAddress(pubKeyProtoType proto.Message) types.Address {
	var pubKey types.PubKey
	err := k.pluginHost.Resolve(pubKeyProtoType, &pubKey)
	if err != nil {
		panic(err)
	}

	return pubKey.Address()
}
