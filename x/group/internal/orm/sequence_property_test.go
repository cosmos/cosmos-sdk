package orm

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

func TestSequence(t *testing.T) {
	rapid.Check(t, rapid.Run[*sequenceMachine]())
}

// sequenceMachine is a state machine model of Sequence. It simply uses a uint64
// as the model of the sequence.
type sequenceMachine struct {
	store sdk.KVStore
	seq   *Sequence
	state uint64
}

// Init sets up the real Sequence, including choosing a random initial value,
// and intialises the model state
func (m *sequenceMachine) Init(t *rapid.T) {
	// Create context and KV store
	ctx := NewMockContext()
	m.store = ctx.KVStore(sdk.NewKVStoreKey("test"))

	// Create primary key table
	seq := NewSequence(0x1)
	m.seq = &seq

	// Choose initial sequence value
	initSeqVal := rapid.Uint64().Draw(t, "initSeqVal")
	err := m.seq.InitVal(m.store, initSeqVal)
	require.NoError(t, err)

	// Create model state
	m.state = initSeqVal
}

// Check does nothing, because all our invariants are captured in the commands
func (m *sequenceMachine) Check(t *rapid.T) {}
