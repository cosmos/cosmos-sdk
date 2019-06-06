package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Operation runs a state machine transition, and ensures the transition
// happened as exported.  The operation could be running and testing a fuzzed
// transaction, or doing the same for a message.
//
// For ease of debugging, an operation returns a descriptive message "action",
// which details what this fuzzed state machine transition actually did.
//
// Operations can optionally provide a list of "FutureOperations" to run later
// These will be ran at the beginning of the corresponding block.
type Operation func(r *rand.Rand, app *baseapp.BaseApp,
	ctx sdk.Context, accounts []Account) (
	OperationMsg OperationMsg, futureOps []FutureOperation, err error)

// entry kinds for use within OperationEntry
const (
	BeginBlockEntryKind  = "begin_block"
	EndBlockEntryKind    = "end_block"
	MsgEntryKind         = "msg"
	QueuedsgMsgEntryKind = "queued_msg"
)

// OperationEntry - an operation entry for logging (ex. BeginBlock, EndBlock, XxxMsg, etc)
type OperationEntry struct {
	EntryKind string          `json:"entry_kind"`
	Height    int64           `json:"height"`
	Order     int64           `json:"order"`
	Operation json.RawMessage `json:"operation"`
}

// BeginBlockEntry - operation entry for begin block
func BeginBlockEntry(height int64) OperationEntry {
	return OperationEntry{
		EntryKind: BeginBlockEntryKind,
		Height:    height,
		Order:     -1,
		Operation: nil,
	}
}

// EndBlockEntry - operation entry for end block
func EndBlockEntry(height int64) OperationEntry {
	return OperationEntry{
		EntryKind: EndBlockEntryKind,
		Height:    height,
		Order:     -1,
		Operation: nil,
	}
}

// MsgEntry - operation entry for standard msg
func MsgEntry(height int64, opMsg OperationMsg, order int64) OperationEntry {
	return OperationEntry{
		EntryKind: MsgEntryKind,
		Height:    height,
		Order:     order,
		Operation: opMsg.MustMarshal(),
	}
}

// MsgEntry - operation entry for queued msg
func QueuedMsgEntry(height int64, opMsg OperationMsg) OperationEntry {
	return OperationEntry{
		EntryKind: QueuedsgMsgEntryKind,
		Height:    height,
		Order:     -1,
		Operation: opMsg.MustMarshal(),
	}
}

// OperationEntry - log entry text for this operation entry
func (oe OperationEntry) MustMarshal() json.RawMessage {
	out, err := json.Marshal(oe)
	if err != nil {
		panic(err)
	}
	return out
}

//_____________________________________________________________________

// OperationMsg - structure for operation output
type OperationMsg struct {
	Route   string          `json:"route"`
	Name    string          `json:"name"`
	Comment string          `json:"comment"`
	OK      bool            `json:"ok"`
	Msg     json.RawMessage `json:"msg"`
}

// OperationMsg - create a new operation message from sdk.Msg
func NewOperationMsg(msg sdk.Msg, ok bool, comment string) OperationMsg {

	return OperationMsg{
		Route:   msg.Route(),
		Name:    msg.Type(),
		Comment: comment,
		OK:      ok,
		Msg:     msg.GetSignBytes(),
	}
}

// OperationMsg - create a new operation message from raw input
func NewOperationMsgBasic(route, name, comment string, ok bool, msg []byte) OperationMsg {
	return OperationMsg{
		Route:   route,
		Name:    name,
		Comment: comment,
		OK:      ok,
		Msg:     msg,
	}
}

// NoOpMsg - create a no-operation message
func NoOpMsg() OperationMsg {
	return OperationMsg{
		Route:   "",
		Name:    "no-operation",
		Comment: "",
		OK:      false,
		Msg:     nil,
	}
}

// log entry text for this operation msg
func (om OperationMsg) String() string {
	out, err := json.Marshal(om)
	if err != nil {
		panic(err)
	}
	return string(out)
}

// Marshal the operation msg, panic on error
func (om OperationMsg) MustMarshal() json.RawMessage {
	out, err := json.Marshal(om)
	if err != nil {
		panic(err)
	}
	return out
}

// add event for event stats
func (om OperationMsg) LogEvent(eventLogger func(string)) {
	pass := "ok"
	if !om.OK {
		pass = "failure"
	}
	eventLogger(fmt.Sprintf("%v/%v/%v", om.Route, om.Name, pass))
}

// queue of operations
type OperationQueue map[int][]Operation

func newOperationQueue() OperationQueue {
	operationQueue := make(OperationQueue)
	return operationQueue
}

// adds all future operations into the operation queue.
func queueOperations(queuedOps OperationQueue,
	queuedTimeOps []FutureOperation, futureOps []FutureOperation) {

	if futureOps == nil {
		return
	}

	for _, futureOp := range futureOps {
		if futureOp.BlockHeight != 0 {
			if val, ok := queuedOps[futureOp.BlockHeight]; ok {
				queuedOps[futureOp.BlockHeight] = append(val, futureOp.Op)
			} else {
				queuedOps[futureOp.BlockHeight] = []Operation{futureOp.Op}
			}
			continue
		}

		// TODO: Replace with proper sorted data structure, so don't have the
		// copy entire slice
		index := sort.Search(
			len(queuedTimeOps),
			func(i int) bool {
				return queuedTimeOps[i].BlockTime.After(futureOp.BlockTime)
			},
		)
		queuedTimeOps = append(queuedTimeOps, FutureOperation{})
		copy(queuedTimeOps[index+1:], queuedTimeOps[index:])
		queuedTimeOps[index] = futureOp
	}
}

//________________________________________________________________________

// FutureOperation is an operation which will be ran at the beginning of the
// provided BlockHeight. If both a BlockHeight and BlockTime are specified, it
// will use the BlockHeight. In the (likely) event that multiple operations
// are queued at the same block height, they will execute in a FIFO pattern.
type FutureOperation struct {
	BlockHeight int
	BlockTime   time.Time
	Op          Operation
}

//________________________________________________________________________

// WeightedOperation is an operation with associated weight.
// This is used to bias the selection operation within the simulator.
type WeightedOperation struct {
	Weight int
	Op     Operation
}

// WeightedOperations is the group of all weighted operations to simulate.
type WeightedOperations []WeightedOperation

func (ops WeightedOperations) totalWeight() int {
	totalOpWeight := 0
	for _, op := range ops {
		totalOpWeight += op.Weight
	}
	return totalOpWeight
}

type selectOpFn func(r *rand.Rand) Operation

func (ops WeightedOperations) getSelectOpFn() selectOpFn {
	totalOpWeight := ops.totalWeight()
	return func(r *rand.Rand) Operation {
		x := r.Intn(totalOpWeight)
		for i := 0; i < len(ops); i++ {
			if x <= ops[i].Weight {
				return ops[i].Op
			}
			x -= ops[i].Weight
		}
		// shouldn't happen
		return ops[0].Op
	}
}
