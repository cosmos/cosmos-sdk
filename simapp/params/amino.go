// +build test_amino

package params

// MakeEncodingConfig creates an EncodingConfig for an amino based test configuration.
func MakeEncodingConfig() EncodingConfig {
	return MakeAminoEncodingConfig()
}
