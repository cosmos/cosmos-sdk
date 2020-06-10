package tx

import "testing"

func TestTxWrapper(t *testing.T) {
	// TODO:
	// - verify that body and authInfo bytes encoded with DefaultTxEncoder and decoded with DefaultTxDecoder can be
	//   retrieved from GetBodyBytes and GetAuthInfoBytes
	// - create a TxWrapper using NewTxWrapper and:
	//   - verify that calling the SetBody results in the correct GetBodyBytes
	//   - verify that calling the SetAuthInfo results in the correct GetAuthInfoBytes and GetPubKeys
	//   - verify no nil panics
}
