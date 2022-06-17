package types

// CurrentFormat is the currently used format for snapshots.
// Essentially, it versions the format of the snapshot data.
// Snapshots using the same format must be identical across all
// nodes for a given height, so this must be bumped when the binary
// snapshot output changes.
// CurrentFormat of 1 is the original format.
// CurrentFormat of 2 serializes the app version in addition to the original snapshot data.
const CurrentFormat uint32 = 2
