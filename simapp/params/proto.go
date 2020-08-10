// +build !test_amino

package params

// MakeEncodingConfig creates an EncodingConfig for an amino based test configuration.
func MakeEncodingConfig() EncodingConfig {
	// TODO switch this to MakeProtoEncodingConfig as soon as proto JSON is ready
	return MakeHybridEncodingConfig()
}
