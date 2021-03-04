package rosetta

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"testing"
)

func TestClientV2(t *testing.T) {
	cdc, ir := MakeCodec()
	c, err := NewClient(&Config{
		Blockchain:        "",
		Network:           "",
		TendermintRPC:     "tcp://localhost:26657",
		GRPCEndpoint:      "localhost:9090",
		Addr:              "",
		Retries:           0,
		Offline:           false,
		Codec:             cdc,
		InterfaceRegistry: ir,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := c.Bootstrap(); err != nil {
		t.Fatal(err)
	}

	var h int64 = 3

	blockTransactionsResponse, err := c.blockTxs(context.TODO(), &h)
	if err != nil {
		t.Fatal(err)
	}

	for _, tx := range blockTransactionsResponse.Transactions {
		t.Logf("hash: %s", tx.TransactionIdentifier.Hash)
		for _, op := range tx.Operations {
			t.Logf("\t name: %s", op.Type)
			t.Logf("\t\t index: %d", op.OperationIdentifier.Index)
			if op.Amount != nil {
				t.Logf("\t\t coin change: %s%s", op.Amount.Value, op.Amount.Currency.Symbol)
			}
			t.Logf("\t\t address: %s", op.Account.Address)
			t.Logf("\t\t meta: %#v", op.Metadata)
		}
	}
}

func TestT(t *testing.T) {

	const addr = "FD41B1F4CA2222B503F448C481873FA91BAE8492495251602A4E60A73C38F944"
	cdc, ir := MakeCodec()
	c, err := NewClient(&Config{
		Blockchain:        "",
		Network:           "",
		TendermintRPC:     "tcp://localhost:26657",
		GRPCEndpoint:      "localhost:9090",
		Addr:              "",
		Retries:           0,
		Offline:           false,
		Codec:             cdc,
		InterfaceRegistry: ir,
	})
	if err != nil {
		t.Fatal(err)
	}
	c.Bootstrap()
	hashBy, err := hex.DecodeString(addr)
	tx, err := c.clientCtx.Client.Tx(context.Background(), hashBy, true)
	t.Logf("%#v", tx.TxResult)

	// ros resp
	rosTx, err := c.GetTx(context.TODO(), addr)
	if err != nil {
		t.Fatal(err)
	}
	x, err := json.Marshal(rosTx)
	t.Logf("%s", x)
}
