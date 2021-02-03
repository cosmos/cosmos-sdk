package v040

import (
	yaml "gopkg.in/yaml.v2"
)

// String returns a human readable string representation of the parameters.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}
