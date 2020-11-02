package sdk

import (
	"context"
	"testing"
)

func TestGetTx(t *testing.T) {
	c, err := NewClient("localhost:9090", "tcp://localhost:26657")
	if err != nil {
		t.Fatal(err)
	}
	tx, err := c.GetTx(context.Background(), "CEC5600E9950971B79D30368ED47ED7C1859E8A048A60AB13469FF01B4E841B0")
	if err != nil {
		t.Fatal(err)
	}
	t.Fatal(tx.String())
}
