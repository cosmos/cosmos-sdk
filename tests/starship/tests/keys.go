package main

import (
	"github.com/docker/distribution/context"
	"google.golang.org/grpc"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"
)

type TestKey struct {
	name    string
	address sdk.AccAddress
	privKey cryptotypes.PrivKey
}

func NewTestKey(name string) *TestKey {
	priv, _, addr := testdata.KeyTestPubAddr()
	return &TestKey{
		name:    name,
		address: addr,
		privKey: priv,
	}
}

type TestKeys []TestKey

// GetKey returns the priv key and address for the given name
func (t *TestKeys) GetKey(name string) (cryptotypes.PrivKey, string) {
	for _, key := range *t {
		if key.name == name {
			return key.privKey, key.address.String()
		}
	}

	return nil, ""
}

// AddKey adds a new key to the TestKeys and return the address
func (t *TestKeys) AddKey(name string) string {
	key := NewTestKey(name)
	*t = append(*t, *key)
	return key.address.String()
}

// GetAccSeqNumber returns the account number and sequence number for the given address
func GetAccSeqNumber(grpcConn *grpc.ClientConn, address string) (uint64, uint64, error) {
	info, err := auth.NewQueryClient(grpcConn).AccountInfo(context.Background(), &auth.QueryAccountInfoRequest{Address: address})
	if err != nil {
		return 0, 0, err
	}
	return info.Info.GetAccountNumber(), info.Info.GetSequence(), nil
}
