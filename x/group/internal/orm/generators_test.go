package orm

import (
	"pgregory.net/rapid"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
)

// genTableModel generates a new table model. At the moment it doesn't
// generate empty strings for Name.
var genTableModel = rapid.Custom(func(t *rapid.T) *testdata.TableModel {
	return &testdata.TableModel{
		Id:       rapid.Uint64().Draw(t, "id").(uint64),
		Name:     rapid.StringN(1, 100, 150).Draw(t, "name").(string),
		Number:   rapid.Uint64().Draw(t, "number ").(uint64),
		Metadata: []byte(rapid.StringN(1, 100, 150).Draw(t, "metadata").(string)),
	}
})
