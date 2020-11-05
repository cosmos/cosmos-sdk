// +build rosetta_integration_tests

package client

import (
	"context"
	"testing"
)

const integrationAddress = "cosmos1wjmt63j4fv9nqda92nsrp2jp2vsukcke4va3pt"

var sendTxBlock int64 = 15

const expectedTxHash = "C58AB818A7DDD8F762051D1B13FBE81AE9E505B959CBBD9B1262B7E391580128"
const expectedBlockHash = "BF12182330951EBDDBD0562683D61C049BD01E0BF049FF4448F8592C56DACCFB"

func TestClient_Balances(t *testing.T) {
	res, err := integrationClient.Balances(context.Background(), integrationAddress, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, b := range res {
		t.Logf("%s", b.String())
	}
}

func TestClient_BlockByHash(t *testing.T) {
	res, txs, err := integrationClient.BlockByHash(context.Background(), expectedBlockHash)
	if err != nil {
		t.Fatal(err)
	}
	if len(txs) != 1 {
		t.Fatal("unexpected number of tx")
	}
	t.Log(res.Block.Hash().String())
	t.Log(txs[0].HexHash)
}

func TestClient_BlockByHeight(t *testing.T) {

}

func TestClient_GetTx(t *testing.T) {
	res, err := integrationClient.GetTx(context.Background(), expectedTxHash)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%#v", res.GetMsgs()[0].String())
}
