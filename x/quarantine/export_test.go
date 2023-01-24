package quarantine

// This file is available only to unit tests and houses functions for doing
// things with private keeper package stuff.

// Expose some private functions so that they can be unit tested.
var (
	ContainsAddress    = containsAddress
	FindAddresses      = findAddresses
	ContainsSuffix     = containsSuffix
	CreateRecordSuffix = createRecordSuffix
)
