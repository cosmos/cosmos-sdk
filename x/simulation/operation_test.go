package simulation

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/stretchr/testify/assert"
)

func TestNewOperationEntry(t *testing.T) {
	t.Run("Testing fo a panic should occur", testNewOperationEntryPanic(t))
	t.Run("Testing for NewOperationMsg", testNewOperationMsg(t))
}

func testNewOperationEntryPanic(t *testing.T) {
	var entries = []string{BeginBlockEntryKind, EndBlockEntryKind, MsgEntryKind, QueuedMsgEntryKind}
	for v := range entries {
		assert.NotPanics(t, func() { NewOperationEntry(entries[v], 12, -1, nil) }, "The code did not panic")
	}
}

func testNewOperationEntry(t *testing.T) {
	var expected = OperationEntry{
		EntryKind: BeginBlockEntryKind,
		Height:    12,
		Order:     -1,
		Operation: nil,
	}

	var actual = NewOperationEntry(BeginBlockEntryKind, 12, -1, nil)
	assert.Equal(t, actual, expected)
}

// TestEntrys tests the different entry kinds Begin and End
func TestEntrys(t *testing.T) {
	t.Run("Testing for all different types of OperationEntrys", func(t *testing.T) {
		t.Run("T", testBeginBlockEntry(t))
		t.Run("T", testEndBlockEntry(t))
	})
}

// testBeginBlockEntry tests a creation of a OperationEntry at the begining stage of block entry.
func testBeginBlockEntry(t *testing.T) {
	var actual = BeginBlockEntry(12)
	var expected = OperationEntry{
		EntryKind: BeginBlockEntryKind,
		Height:    12,
		Order:     -1,
		Operation: nil,
	}

	assert.Equal(t, actual, expected)
}

// testEndBlockEntry tests creates a OperationEntry at the ending stage of a block entry.
func testEndBlockEntry(t *testing.T) {
	var actual = EndBlockEntry(12)
	var expected = OperationEntry{
		EntryKind: EndBlockEntryKind,
		Height:    12,
		Order:     -1,
		Operation: nil,
	}

	assert.Equal(t, expected, actual)
}

func TestMsgEntry(t *testing.T) {
	var (
		addr         = secp256k1.GenPrivKey().PubKey().Address()
		accAddr      = types.AccAddress(addr)
		msg          = sdk.NewTestMsg(accAddr)
		om           = NewOperationMsg(msg, true, "test")
		omMarshalled = om.MustMarshal()
	)

	var expected = OperationEntry{EntryKind: MsgEntryKind, Height: 1, Order: 1, Operation: omMarshalled}
	var actual = MsgEntry(1, 1, om)
	assert.Equal(t, actual, expected)
}

func TestNewOperationMsgBasic(t *testing.T) {
	var (
		route   = "home"
		name    = "name"
		comment = "comment"
		ok      = true
		msg     = []byte("test")
	)

	var actual = NewOperationMsgBasic(route, name, comment, ok, msg)
	var expected = OperationMsg{Route: route, Name: name, Comment: comment, OK: ok, Msg: msg}

	assert.Equal(t, actual, expected)
}

func TestNoOpMsg(t *testing.T) {
	var actual = NoOpMsg("test")
	var expected = OperationMsg{Route: "test", Name: "no-operation", Comment: "", OK: false, Msg: nil}

	assert.Equal(t, actual, expected)
}

// TestOperationMsgMethods tests every method of OperationMsg
func TestOperationMsgMethods(t *testing.T) {
	t.Run("Testing OperationMsg and its methods", func(t *testing.T) {
		t.Run("NewOpMsg", testNewOpMsg(t))
		t.Run("NewOperationMsg", testNewOperationMsg(t))
		t.Run("OpMsgString", testOpMsgString(t))
		t.Run("OpMustMarshal", testOpMustMarshal(t))
		t.Run("OpMsgLogEvent", testOpMsgLogEvent(t))
	})
}

// testNewOperationMsg tests NewOperationMsg
func testNewOperationMsg(t *testing.T) {
	var (
		addr    = secp256k1.GenPrivKey().PubKey().Address()
		accAddr = types.AccAddress(addr)
		msg     = sdk.NewTestMsg(accAddr)
	)

	var sortedMarshal = func(signers ...types.AccAddress) ([]byte, error) {
		bz, err := json.Marshal(signers)
		if err != nil {
			return nil, err
		}

		return sdk.MustSortJSON(bz), nil
	}

	var expectedSortedMsg, _ = sortedMarshal(accAddr)

	var expected = OperationMsg{Route: "TestMsg", Name: "Test message", Comment: "test", OK: true, Msg: expectedSortedMsg}
	var actual = NewOperationMsg(msg, true, "test")

	assert.Equal(t, expected, actual)
}

// testNewOperationMsg tests NewOpMsg
func testNewOpMsg(t *testing.T) {
	var expected = OperationMsg{Route: "TestMsg", Name: "no-operation", Comment: "", OK: false, Msg: nil}
	var actual = NoOpMsg("TestMsg")

	assert.Equal(t, expected, actual)
}

// getOpMsg is used to return a OperationMsg with a valid msg
var getOpMsg = func() OperationMsg {
	var (
		addr    = secp256k1.GenPrivKey().PubKey().Address()
		accAddr = types.AccAddress(addr)
		msg     = sdk.NewTestMsg(accAddr)
	)

	return NewOperationMsg(msg, true, "test")
}

// testOpMsgString tests OpMsgString
func testOpMsgString(t *testing.T) {
	var opMsg = getOpMsg()

	var expected = func(OperationMsg) string {
		out, _ := json.Marshal(opMsg)

		return string(out)
	}(opMsg)

	var actual = opMsg.String()
	assert.Equal(t, expected, actual)
}

// testOpMustMarshal tests OpMustMarshal
func testOpMustMarshal(t *testing.T) {
	var opMsg = getOpMsg()

	var expected = func(OperationMsg) json.RawMessage {
		out, _ := json.Marshal(opMsg)
		return out
	}(opMsg)

	var actual = opMsg.MustMarshal()

	assert.Equal(t, expected, actual)
}

// testOpMsgLogEvent tests OpMsgLogEvent
func testOpMsgLogEvent(t *testing.T) {
	var (
		route, ep          = "TestMsg", "Test message"
		createExpectedLogs = func(s bool) {
			f, _ := os.Create("expectedLogs")
			var success string
			if s == true {
				success = "ok"
			} else {
				success = "nonok"
			}
			_, _ = f.WriteString(route + ":" + ep + ":" + success)
		}
		eventLogger = func(route, ep, evResult string) {
			f, _ := os.Create("actualLogs")
			_, _ = f.WriteString(route + ":" + ep + ":" + evResult)
			return
		}
		resetLogs = func() { os.Remove("expectedLogs"); os.Remove("actualLogs") }
	)

	var tests = []struct {
		description string
		success     bool
		expected    string
	}{
		{"Case A: opMsg.OK=false", false, "nonok"},
		{"Case B: opMsg.OK=true", true, "ok"},
	}

	for _, v := range tests {
		t.Run(fmt.Sprintf("Testing: %v with %v and %v", v.description, v.expected), func(t *testing.T) {
			createExpectedLogs(v.success)
			var expected, _ = ioutil.ReadFile("expectedLogs")

			var opMsg = getOpMsg()

			opMsg.OK = v.success
			opMsg.LogEvent(eventLogger)
			var actual, _ = ioutil.ReadFile("actualLogs")

			assert.Equal(t, actual, expected)
			resetLogs()
		})
	}
}

var genOp = func() Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp,
		ctx sdk.Context, accounts []Account, chainID string) {
	}
}

func randomTimestamp() time.Time {
	randomTime := rand.Int63n(time.Now().Unix()-94608000) + 94608000
	randomNow := time.Unix(randomTime, 0)

	return randomNow
}

var newFutureOpt = func(rand bool, t string) FutureOperation {
	fo := new(FutureOperation)
	fo.Op = genOp()

	if rand {
		switch rand.Intn(3) {
		case 1:
			fo.BlockHeight = rand.Intn(4000)
			return fo
		case 2:
			fo.BlockTime = randomTimestamp()
			return fo
		case 3:
			fo.BlockHeight = rand.Intn(4000)
			fo.BlockTime = randomTimestamp()
			return fo
		}
	}

	switch t {
	case "height":
		fo.BlockHeight = rand.Intn(4000)
		return fo
	case "blockTime":
		fo.BlockTime = randomTimestamp()
		return fo
	default:
		fo.BlockHeight = rand.Intn(4000)
		fo.BlockTime = randomTimestamp()
		return fo
	}
}

var createFutureOps = func(s string) []FutureOperation {
	r1 := rand.Seed(time.Now().UnixNano()).Int()
	futureOpts := make([]FutureOperation, r1)

	switch s {
	case "Operations with block heights":
		for i := 0; i < len(futureOpts); i++ {
			futureOpts = append(futureOpts, newFutureOpt(false, "height"))
		}
	case "Operations with block times":
		for i := 0; i < len(futureOpts); i++ {
			futureOpts = append(futureOpts, newFutureOpt(false, "blockTime"))
		}

		return futureOpts
	case "Operations with block times and heights":
		for i := 0; i < len(futureOpts); i++ {
			futureOpts = append(futureOpts, newFutureOpt(true, ""))
		}
		return futureOpts
	default:
		return nil

	}
}

type queueOptsCases struct {
	description      string
	queuedTimeOps    []FutureOperation
	queuedFutureOps  []FutureOperation
	queuedOperations OperationQueue

	expectationFutureOpt func([]FutureOperation) bool
	expectationQueuedOps func(OperationQueue, ...[]FutureOperation) bool
}

// Test strategy:
// 1. Test if futureOpts is equal to zero i.e it has been drained of its operations
// 2. Test if queuedTimeOps is properly sorted and only has blockheight == 0
// 3. Benchmark test between A and B ensure that B is faster then A
// 4. Test that the end len(queuedTimeOps) + len(futureOpts) + len(opeartionQueue) == beginning
func testQueueOperationsA(t *testing.T) {
	queueOpTest := []queueOptsCases{
		{
			description:      "Test correct queueing of operations with only block heights, operations with heights should not be queued within FutureTimeOpts",
			queuedFutureOps:  createFutureOps("Operation with block heights"),
			queuedTimeOps:    make([]FutureOperation, 0),
			queuedOperations: NewOperationQueue(),
			expectationFutureOpt: func(fto []FutureOperation) bool {
				for _, v := range fto {
					if v.BlockHeight != 0 {
						return false
					}
				}

				return true
			},
		},
		{
			description:      "Tests the draining of the original []FutureOperation queue, all operations should of been dispersed.",
			queuedFutureOps:  createFutureOps(""),
			queuedTimeOps:    make([]FutureOperation, 0),
			queuedOperations: NewOperationQueue(),
			expectationFutureOpt: func(fo []FutureOperation) bool {
				if len(fo) != 0 {
					return false
				}

				return true
			},
		},
		{
			description:      "Test that futureTimeOpts only has operations with block times",
			queuedFutureOps:  createFutureOps(""),
			queuedTimeOps:    make([]FutureOperation, 0),
			queuedOperations: NewOperationQueue(),
			expectationFutureOpt: func(fo []FutureOperation) bool {
				for _, v := range fo {
					if v.BlockTime != (time.Time{}) {
						return false
					}
				}

				return true
			},
		},
		{
			description:      "Test that all operations were processed. Total size of operation queue should equal the size of both future operation queues.",
			queuedFutureOps:  createFutureOps(""),
			queuedTimeOps:    make([]FutureOperation, 0),
			queuedOperations: NewOperationQueue(),
			expectationQueuedOps: func(oq OperationQueue, fo ...[]FutureOperation) bool {
				size := 0
				for _, v := range fo {
					size += len(v)
				}

				if len(oq) != size {
					return false
				}

				return true
			},
		},
	}

	for _, o := range queueOpTest {
		t.Run(fmt.Sprintf("Testing: %v", o.description), func(t *testing.T) {
			queueOperations(o.queuedOperations, o.queuedTimeOps, o.queuedFutureOps)
			assert.Condition(t, o.expectationFutureOpt(o.queuedOperations))
		})
	}
}
