package indexertesting

import indexerbase "cosmossdk.io/indexer/base"

// ListenerTestFixture is a test fixture for testing listener implementations with a pre-defined data set
// that attempts to cover all known types of tables and columns. The test data currently includes data for
// two fake modules over three blocks of data. The data set should remain relatively stable between releases
// and generally only be changed when new features are added, so it should be suitable for regression or golden tests.
type ListenerTestFixture struct {
	listener indexerbase.Listener
}

type ListenerTestFixtureOptions struct {
	EventAlignedWrites bool
}

func NewListenerTestFixture(listener indexerbase.Listener, options ListenerTestFixtureOptions) *ListenerTestFixture {
	return &ListenerTestFixture{
		listener: listener,
	}
}

func (f *ListenerTestFixture) Initialize() error {
	return nil
}

func (f *ListenerTestFixture) NextBlock() (bool, error) {
	return false, nil
}

func (f *ListenerTestFixture) block1() error {
	return nil
}

func (f *ListenerTestFixture) block2() error {
	return nil
}

func (f *ListenerTestFixture) block3() error {
	return nil
}

var moduleSchemaA = indexerbase.ModuleSchema{
	Tables: []indexerbase.Table{
		{
			"A1",
			[]indexerbase.Column{},
			[]indexerbase.Column{},
			false,
		},
	},
}

var moduleSchemaB = indexerbase.ModuleSchema{
	Tables: []indexerbase.Table{
		{
			"B1",
			[]indexerbase.Column{},
			[]indexerbase.Column{},
			false,
		},
	},
}
