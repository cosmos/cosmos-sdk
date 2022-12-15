package orm

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
)

func TestPrimaryKeyTable(t *testing.T) {
	rapid.Check(t, rapid.Run[*primaryKeyMachine]())
}

// primaryKeyMachine is a state machine model of the PrimaryKeyTable. The state
// is modelled as a map of strings to TableModels.
type primaryKeyMachine struct {
	store sdk.KVStore
	table *PrimaryKeyTable
	state map[string]*testdata.TableModel
}

// stateKeys gets all the keys in the model map
func (m *primaryKeyMachine) stateKeys() []string {
	keys := make([]string, len(m.state))

	i := 0
	for k := range m.state {
		keys[i] = k
		i++
	}

	return keys
}

// Init creates a new instance of the state machine model by building the real
// table and making the empty model map
func (m *primaryKeyMachine) Init(t *rapid.T) {
	// Create context
	ctx := NewMockContext()
	m.store = ctx.KVStore(sdk.NewKVStoreKey("test"))

	// Create primary key table
	interfaceRegistry := types.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)
	table, err := NewPrimaryKeyTable(
		[2]byte{0x1},
		&testdata.TableModel{},
		cdc,
	)
	require.NoError(t, err)

	m.table = table

	// Create model state
	m.state = make(map[string]*testdata.TableModel)
}

// Check that the real values match the state values.
func (m *primaryKeyMachine) Check(t *rapid.T) {
	for i := range m.state {
		has := m.table.Has(m.store, []byte(i))
		require.Equal(t, true, has)
	}
}
