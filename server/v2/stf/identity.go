package stf

var (
	// Identity defines STF's bytes identity and it's used by STF to store things in its own state.
	Identity = []byte("stf")
	// RuntimeIdentity defines the bytes identity of the runtime.
	RuntimeIdentity = []byte("runtime")
	// ConsensusIdentity defines the bytes identity of the consensus.
	ConsensusIdentity = []byte("consensus")
)
