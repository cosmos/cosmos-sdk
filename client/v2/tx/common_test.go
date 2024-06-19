package tx

import (
	"context"
	abciv1beta1 "cosmossdk.io/api/cosmos/base/abci/v1beta1"
	apitx "cosmossdk.io/api/cosmos/tx/v1beta1"
	"github.com/cosmos/cosmos-sdk/types"

	"google.golang.org/grpc"

	"cosmossdk.io/client/v2/autocli/keyring"
	txdecode "cosmossdk.io/x/tx/decode"
	"cosmossdk.io/x/tx/signing"

	"github.com/cosmos/cosmos-sdk/codec"
	addrcodec "github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	codec2 "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	cryptoKeyring "github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

var (
	cdc            = codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
	ac             = addrcodec.NewBech32Codec("cosmos")
	valCodec       = addrcodec.NewBech32Codec("cosmosval")
	signingOptions = signing.Options{
		AddressCodec:          ac,
		ValidatorAddressCodec: valCodec,
	}
	signingContext, _ = signing.NewContext(signingOptions)
	decodeOptions     = txdecode.Options{SigningContext: signingContext, ProtoCodec: cdc}
	decoder, _        = txdecode.NewDecoder(decodeOptions)

	k          = cryptoKeyring.NewInMemory(cdc)
	keybase, _ = cryptoKeyring.NewAutoCLIKeyring(k, ac)
	txConf, _  = NewTxConfig(ConfigOptions{
		AddressCodec:          ac,
		Cdc:                   cdc,
		ValidatorAddressCodec: valCodec,
	})
)

func setKeyring() keyring.Keyring {
	registry := codectypes.NewInterfaceRegistry()
	codec2.RegisterInterfaces(registry)
	cdc := codec.NewProtoCodec(registry)
	k := cryptoKeyring.NewInMemory(cdc)
	_, err := k.NewAccount("alice", "equip will roof matter pink blind book anxiety banner elbow sun young", "", "m/44'/118'/0'/0/0", hd.Secp256k1)
	if err != nil {
		panic(err)
	}
	keybase, err := cryptoKeyring.NewAutoCLIKeyring(k, ac)
	if err != nil {
		panic(err)
	}
	return keybase
}

type mockAccount struct {
	addr []byte
}

func (m mockAccount) GetAddress() types.AccAddress {
	return m.addr
}

func (m mockAccount) GetPubKey() cryptotypes.PubKey {
	return nil
}

func (m mockAccount) GetAccountNumber() uint64 {
	return 1
}

func (m mockAccount) GetSequence() uint64 {
	return 0
}

type mockAccountRetriever struct{}

func (m mockAccountRetriever) GetAccount(_ context.Context, address []byte) (Account, error) {
	return mockAccount{addr: address}, nil
}

func (m mockAccountRetriever) GetAccountWithHeight(_ context.Context, address []byte) (Account, int64, error) {
	return mockAccount{addr: address}, 0, nil
}

func (m mockAccountRetriever) EnsureExists(_ context.Context, address []byte) error {
	return nil
}

func (m mockAccountRetriever) GetAccountNumberSequence(_ context.Context, address []byte) (accNum, accSeq uint64, err error) {
	return accNum, accSeq, nil
}

type mockClientConn struct{}

func (m mockClientConn) Invoke(_ context.Context, _ string, args, reply interface{}, opts ...grpc.CallOption) error {
	simResponse := apitx.SimulateResponse{
		GasInfo: &abciv1beta1.GasInfo{ // TODO: sdk dependency
			GasWanted: 10000,
			GasUsed:   7500,
		},
		Result: nil,
	}
	*reply.(*apitx.SimulateResponse) = simResponse
	return nil
}

func (m mockClientConn) NewStream(_ context.Context, _ *grpc.StreamDesc, _ string, _ ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, nil
}
