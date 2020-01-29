package simulation

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/cosmos/cosmos-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
)

func TestNewOperationEntry(t *testing.T) {
	var entries = []string{BeginBlockEntryKind, EndBlockEntryKind, MsgEntryKind, QueuedMsgEntryKind}

	// test fail if does not contain one of the entrys
	for v := range entries {
		assert.NotPanics(t, func() { NewOperationEntry(entries[v], 12, -1, nil) }, "The code did not panic")
	}

	// test for new operation entry
	var expected = OperationEntry{
		EntryKind: BeginBlockEntryKind,
		Height:    12,
		Order:     -1,
		Operation: nil,
	}

	var actual = NewOperationEntry(BeginBlockEntryKind, 12, -1, nil)
	if ok := reflect.DeepEqual(expected, actual); !ok {
		t.Fatalf("\nactual: %v\n expected: %v", spew.Sprint(actual), spew.Sprint(expected))
	}
}

// testBeginBlockEntry tests a creation of a OperationEntry at the begining stage of block entry.
func TestBeginBlockEntry(t *testing.T) {
	var actual = BeginBlockEntry(12)
	var expected = OperationEntry{
		EntryKind: BeginBlockEntryKind,
		Height:    12,
		Order:     -1,
		Operation: nil,
	}

	if ok := reflect.DeepEqual(expected, actual); !ok {
		t.Fatalf("\nactual: %v\n expected: %v", spew.Sprint(actual), spew.Sprint(expected))
	}
}

// testEndBlockEntry tests creates a OperationEntry at the ending stage of a block entry.
func TestEndBlockEntry(t *testing.T) {
	var actual = EndBlockEntry(12)
	var expected = OperationEntry{
		EntryKind: EndBlockEntryKind,
		Height:    12,
		Order:     -1,
		Operation: nil,
	}

	if ok := reflect.DeepEqual(expected, actual); !ok {
		t.Fatalf("\nactual: %v\n expected: %v", spew.Sprint(actual), spew.Sprint(expected))
	}
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
	if ok := reflect.DeepEqual(expected, actual); !ok {
		t.Fatalf("\nactual: %v\n expected: %v", spew.Sprint(actual), spew.Sprint(expected))
	}
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

	if ok := reflect.DeepEqual(expected, actual); !ok {
		t.Fatalf("\nactual: %v\n expected: %v", spew.Sprint(actual), spew.Sprint(expected))
	}
}

func TestNoOpMsg(t *testing.T) {
	var actual = NoOpMsg("test")

	var expected = OperationMsg{Route: "test", Name: "no-operation", Comment: "", OK: false, Msg: nil}

	if ok := reflect.DeepEqual(expected, actual); !ok {
		t.Fatalf("\nactual: %v\n expected: %v", spew.Sprint(actual), spew.Sprint(expected))
	}
}

func TestOperationMsgMethods(t *testing.T) {

	testNewOpMsg(t)
	testNewOperationMsg(t)
	testOpMsgString(t)
	testOpMustMarshal(t)
	testOpMsgLogEvent(t)

	// 	testOperationMsgString(t, om)
}

func testNewOperationMsg(t *testing.T) OperationMsg {
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

	if ok := reflect.DeepEqual(expected, actual); !ok {
		t.Fatalf("\nactual: %v\n expected: %v", spew.Sprint(actual), spew.Sprint(expected))
	}

	return actual
}

func testNewOpMsg(t *testing.T) {
	var expected = OperationMsg{Route: "TestMsg", Name: "no-operation", Comment: "", OK: false, Msg: nil}
	var actual = NoOpMsg("TestMsg")

	if ok := reflect.DeepEqual(expected, actual); !ok {
		t.Fatalf("\nactual: %v\n expected: %v", spew.Sprint(actual), spew.Sprint(expected))
	}
}

var getOpMsg = func() OperationMsg {
	var (
		addr    = secp256k1.GenPrivKey().PubKey().Address()
		accAddr = types.AccAddress(addr)
		msg     = sdk.NewTestMsg(accAddr)
	)

	return NewOperationMsg(msg, true, "test")
}

func testOpMsgString(t *testing.T) {
	var (
		opMsg = getOpMsg()
	)

	var expected = func(OperationMsg) string {
		out, _ := json.Marshal(opMsg)

		return string(out)
	}(opMsg)

	var actual = opMsg.String()

	if ok := reflect.DeepEqual(expected, actual); !ok {
		t.Fatalf("\nactual: %v\n expected: %v", spew.Sprint(actual), spew.Sprint(expected))
	}
}

func testOpMustMarshal(t *testing.T) {
	var (
		opMsg = getOpMsg()
	)

	var expected = func(OperationMsg) json.RawMessage {
		out, _ := json.Marshal(opMsg)
		return out
	}(opMsg)

	var actual = opMsg.MustMarshal()

	if ok := reflect.DeepEqual(expected, actual); !ok {
		t.Fatalf("\nactual: %v\n expected: %v", spew.Sprint(actual), spew.Sprint(expected))
	}
}

func testOpMsgLogEvent(t *testing.T) {
	var (
		opMsg              = getOpMsg()
		route, ep          = "TestMsg", "Test message"
		createExpectedLogs = func(s string) {
			f, _ := os.Create("expectedLogs")
			_, _ = f.WriteString(route + ":" + ep + ":" + s)
		}

		el = func(route, ep, evResult string) {
			f, _ := os.Create("actualLogs")
			_, _ = f.WriteString(route + ":" + ep + ":" + evResult)

			return
		}

		resetLogs = func() {
			os.Remove("expectedLogs")
			os.Remove("actualLogs")
		}
	)

	// test failure event
	createExpectedLogs("failure")
	var expected, _ = ioutil.ReadFile("expectedLogs")

	opMsg.OK = false
	opMsg.LogEvent(el)

	var actual, _ = ioutil.ReadFile("actualLogs")
	err := bytes.Compare(actual, expected)
	if err != 0 {
		resetLogs()
		t.Fatalf("\nactual: %v\n expected: %v", spew.Sprint(string(actual)), spew.Sprint(string(expected)))
	}

	// test non-failure event
	createExpectedLogs("ok")
	expected, _ = ioutil.ReadFile("expectedLogs")

	opMsg.OK = true
	opMsg.LogEvent(el)
	actual, _ = ioutil.ReadFile("actualLogs")
	err = bytes.Compare(actual, expected)
	if err != 0 {
		resetLogs()
		t.Fatalf("\nactual: %v\n expected: %v", spew.Sprint(string(actual)), spew.Sprint(string(expected)))
	}

	// kill test logs
	resetLogs()
}

/*
func TestQueueOperations(t *testing.T) {

	var (
		queuedTimeOps =
		operationQueue = NewOperationQueue()
		futureOps =
	)

	queueOpeartions(queuedOps OpeartionQueue, queuedTimesOps, futureOps)
}
*/
