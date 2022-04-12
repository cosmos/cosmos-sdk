package multi

import (
	"testing"
)

func TestV1asV2MultiStoreBasic(t *testing.T) {
	doTestMultiStoreBasic(t, NewV1MultiStoreAsV2)
}

func TestV1asV2GetVersion(t *testing.T) {
	doTestGetVersion(t, NewV1MultiStoreAsV2)
}

func TestV1asV2Pruning(t *testing.T) {
	doTestPruning(t, NewV1MultiStoreAsV2, false)
}

func TestV1asV2Trace(t *testing.T) {
	doTestTrace(t, NewV1MultiStoreAsV2)
}

func TestV1asV2TraceConcurrency(t *testing.T) {
	doTestTraceConcurrency(t, NewV1MultiStoreAsV2)
}

func TestV1asV2Listeners(t *testing.T) {
	doTestListeners(t, NewV1MultiStoreAsV2)
}
