package indexertesting

import (
	rand "math/rand/v2"

	indexerbase "cosmossdk.io/indexer/base"
)

// ListenerTestFixture is a test fixture for testing listener implementations with a pre-defined data set
// that attempts to cover all known types of objects and fields. The test data currently includes data for
// two fake modules over three blocks of data. The data set should remain relatively stable between releases
// and generally only be changed when new features are added, so it should be suitable for regression or golden tests.
type ListenerTestFixture struct {
	rndSource rand.Source
	block     uint64
	listener  indexerbase.Listener
	//allKeyModule *testModule
}

type ListenerTestFixtureOptions struct {
	EventAlignedWrites bool
}
