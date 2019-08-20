package simulation

import (
	"encoding/json"
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
	BeginBlockEntryKind = "begin_block"
	EndBlockEntryKind   = "end_block"
	MsgEntryKind        = "msg"
	QueuedMsgEntryKind  = "queued_msg"
)

// OperationEntry - an operation entry for logging (ex. BeginBlock, EndBlock, XxxMsg, etc)
type OperationEntry struct {
	EntryKind string          `json:"entry_kind" yaml:"entry_kind"`
	Height    int64           `json:"height" yaml:"height"`
	Order     int64           `json:"order" yaml:"order"`
	Operation json.RawMessage `json:"operation" yaml:"operation"`
}

// NewOperationEntry creates a new OperationEntry instance
func NewOperationEntry(entry string, height, order int64, op json.RawMessage) OperationEntry {
	return OperationEntry{
		EntryKind: entry,
		Height:    height,
		Order:     order,
		Operation: op,
	}
}

// BeginBlockEntry - operation entry for begin block
func BeginBlockEntry(height int64) OperationEntry {
	return NewOperationEntry(BeginBlockEntryKind, height, -1, nil)
}

// EndBlockEntry - operation entry for end block
func EndBlockEntry(height int64) OperationEntry {
	return NewOperationEntry(EndBlockEntryKind, height, -1, nil)
}

// MsgEntry - operation entry for standard msg
func MsgEntry(height, order int64, opMsg OperationMsg) OperationEntry {
	return NewOperationEntry(MsgEntryKind, height, order, opMsg.MustMarshal())
}

// QueuedMsgEntry creates an operation entry for a given queued message.
func QueuedMsgEntry(height int64, opMsg OperationMsg) OperationEntry {
	return NewOperationEntry(QueuedMsgEntryKind, height, -1, opMsg.MustMarshal())
}

// MustMarshal marshals the operation entry, panic on error.
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
	Route   string          `json:"route" yaml:"route"`
	Name    string          `json:"name" yaml:"name"`
	Comment string          `json:"comment" yaml:"comment"`
	OK      bool            `json:"ok" yaml:"ok"`
	Msg     json.RawMessage `json:"msg" yaml:"msg"`
}

// NewOperationMsgBasic creates a new operation message from raw input.
func NewOperationMsgBasic(route, name, comment string, ok bool, msg []byte) OperationMsg {
	return OperationMsg{
		Route:   route,
		Name:    name,
		Comment: comment,
		OK:      ok,
		Msg:     msg,
	}
}

// NewOperationMsg - create a new operation message from sdk.Msg
func NewOperationMsg(msg sdk.Msg, ok bool, comment string) OperationMsg {
	return NewOperationMsgBasic(msg.Route(), msg.Type(), comment, ok, msg.GetSignBytes())
}

// NoOpMsg - create a no-operation message
func NoOpMsg(route string) OperationMsg {
	return NewOperationMsgBasic(route, "no-operation", "", false, nil)
}

// log entry text for this operation msg
func (om OperationMsg) String() string {
	out, err := json.Marshal(om)
	if err != nil {
		panic(err)
	}
	return string(out)
}

// MustMarshal Marshals the operation msg, panic on error
func (om OperationMsg) MustMarshal() json.RawMessage {
	out, err := json.Marshal(om)
	if err != nil {
		panic(err)
	}
	return out
}

// LogEvent adds an event for the events stats
func (om OperationMsg) LogEvent(eventLogger func(route, op, evResult string)) {
	pass := "ok"
	if !om.OK {
		pass = "failure"
	}
	eventLogger(om.Route, om.Name, pass)
}

// OperationQueue defines an object for a queue of operations
type OperationQueue map[int][]Operation

// NewOperationQueue creates a new OperationQueue instance.
func NewOperationQueue() OperationQueue {
	return make(OperationQueue)
}

// queueOperations adds all future operations into the operation queue.
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
