package offchain

import (
	"testing"
)

func TestSigner_Sign(t *testing.T) {
	tx, err := testSigner.Sign(testPrivKey, []msg{
		NewMsgSignData(testAddress, []byte("data")),
	})
	if err != nil {
		t.Fatal(err)
	}
	// verify omfg
	err = testVerifier.Verify(tx)
	if err != nil {
		t.Fatal(err)
	}
}
