package reflection

import (
	"context"
	"testing"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/cosmos/cosmos-sdk/client/grpc/reflection"
	"github.com/cosmos/cosmos-sdk/client/reflection/unstructured"
)

func TestClient(t *testing.T) {
	c, err := NewClient("localhost:9090", "")
	if err != nil {
		t.Fatal(err)
	}

	qs, err := c.sdkReflect.ListQueryServices(context.TODO(), nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%s", qs)

	imp, err := c.sdkReflect.ListDeliverables(context.TODO(), nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%s", imp)

	typeDesc, err := c.sdkReflect.ResolveProtoType(context.TODO(), &reflection.ResolveProtoTypeRequest{Name: "cosmos.bank.v1beta1.MsgSend"})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(typeDesc)

	svcDesc, err := c.sdkReflect.ResolveService(context.TODO(), &reflection.ResolveServiceRequest{FileName: qs.Queries[1].ProtoFile})
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%s", svcDesc)
}

func TestClientListQueries(t *testing.T) {
	c, err := NewClient("localhost:9090", "")
	if err != nil {
		t.Fatal(err)
	}

	qs := c.ListQueries()
	for _, q := range qs {
		t.Log(q.String())
	}
}

func TestClient_Query(t *testing.T) {
	c, err := NewClient("localhost:9090", "tcp://localhost:26657")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("bank test", func(t *testing.T) {
		resp, err := c.Query(context.TODO(), "/cosmos.bank.v1beta1.Query/Balance", unstructured.Object{
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
		resp, err := c.Query(context.TODO(), "/cosmos.bank.v1beta1.Query/Params", unstructured.Object{})
		if err != nil {
			t.Fatal(err)
		}
		t.Log(resp)
	})

}
