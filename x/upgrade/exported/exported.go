package exported

// ProtocolVersionSetter defines the interface fulfilled by BaseApp
// which allows setting its appVersion field.
type ProtocolVersionSetter interface {
	SetProtocolVersion(uint64)
}
