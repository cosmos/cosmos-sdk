package client

import (
	"context"
	"encoding/hex"
	"testing"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/cosmos/cosmos-sdk/client/grpc/reflection"
	"github.com/cosmos/cosmos-sdk/client/reflection/tx"
	"github.com/cosmos/cosmos-sdk/client/reflection/unstructured"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
)

func TestClient(t *testing.T) {
	c, err := DialContext(context.TODO(), "localhost:9090", "", nil)
	if err != nil {
		t.Fatal(err)
	}

	qs, err := c.sdk.ListQueryServices(context.TODO(), nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%s", qs)

	imp, err := c.sdk.ListDeliverables(context.TODO(), nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%s", imp)

	typeDesc, err := c.sdk.ResolveProtoType(context.TODO(), &reflection.ResolveProtoTypeRequest{Name: "cosmos.bank.v1beta1.MsgSend"})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(typeDesc)

	svcDesc, err := c.sdk.ResolveService(context.TODO(), &reflection.ResolveServiceRequest{FileName: qs.Queries[1].ProtoFile})
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%s", svcDesc)
}

func TestClientListQueries(t *testing.T) {

}

func TestClient_ListDeliverables(t *testing.T) {

}

func TestClient_resolveAnys(t *testing.T) {
	c, err := DialContext(context.TODO(), "localhost:9090", "", nil)
	if err != nil {
		t.Fatal(err)
	}

	err = c.resolveAnys(context.TODO())
	if err != nil {
		t.Fatal(err)
	}
}

func TestClient_Query(t *testing.T) {
	c, err := DialContext(context.TODO(), "localhost:9090", "tcp://localhost:26657", nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("account test", func(t *testing.T) {
		resp, err := c.QueryUnstructured(context.TODO(), "/cosmos.auth.v1beta1.Query/Account", unstructured.Map{
			"address": "cosmos1ujtnemf6jmfm995j000qdry064n5lq854gfe3j",
		})
		if err != nil {
			t.Fatal(err)
		}

		b, err := c.cdc.MarshalJSON(resp)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%s", b)
	})

	t.Run("bank test", func(t *testing.T) {
		resp, err := c.QueryUnstructured(context.TODO(), "/cosmos.bank.v1beta1.Query/Balance", unstructured.Map{
			"address": "cosmos1ujtnemf6jmfm995j000qdry064n5lq854gfe3j",
			"denom":   "stake",
		})

		if err != nil {
			t.Fatal(err)
		}

		t.Log(resp)

		b, err := protojson.Marshal(resp)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%s", b)
	})

	t.Run("params", func(t *testing.T) {
		resp, err := c.QueryUnstructured(context.TODO(), "/cosmos.bank.v1beta1.Query/Params", unstructured.Map{})
		if err != nil {
			t.Fatal(err)
		}
		t.Log(resp)
	})

}

type testInfoProvider struct {
	pk            cryptotypes.PrivKey
	sequence      uint64
	accountNumber uint64
}

func newInfoProvider(hexKey string, sequence, accountNumber uint64) testInfoProvider {
	b, err := hex.DecodeString(hexKey)
	if err != nil {
		panic(err)
	}

	pk := &secp256k1.PrivKey{Key: b}
	return testInfoProvider{
		pk:            pk,
		sequence:      sequence,
		accountNumber: accountNumber,
	}
}

func (t testInfoProvider) SigningInfo(ctx context.Context, pubKey cryptotypes.PubKey) (accountNumber, sequence uint64, err error) {
	panic("implement me")
}

func (t testInfoProvider) Sign(ctx context.Context, pubKey cryptotypes.PubKey, b []byte) (signedBytes []byte, err error) {
	return t.pk.Sign(b)
}

func (t testInfoProvider) Address(ctx context.Context, pubKey cryptotypes.PubKey) (address string, err error) {
	panic("implement mwe")
}

func TestClient_Tx(t *testing.T) {
	const keyHex = "8c7e006440ac5e358739bdc3d10a8b2d229e23d27660f6d3a8306cee4379594c"
	const sequence uint64 = 3
	const accNum uint64 = 0
	infoProvider := newInfoProvider(keyHex, 0, 0)
	c, err := DialContext(context.TODO(), "localhost:9090", "tcp://localhost:26657", infoProvider)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.Tx(context.TODO(), "cosmos.bank.v1beta1.MsgSend",
		unstructured.Map{
			"from_address": "cosmos1ujtnemf6jmfm995j000qdry064n5lq854gfe3j",
			"to_address":   "cosmos1caa3es6q3mv8t4gksn9wjcwyzw7cnf5gn5cx7j",
			"amount": []unstructured.Map{
				{
					"denom":  "stake",
					"amount": "10",
				},
			},
		},
		tx.SignerInfo{
			PubKey:        infoProvider.pk.PubKey(),
			SignMode:      signing.SignMode_SIGN_MODE_DIRECT,
			AccountNumber: accNum,
			Sequence:      sequence,
		})

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%#v", resp)
}
