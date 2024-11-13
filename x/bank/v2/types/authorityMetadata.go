package types

import (
	"bytes"
	"errors"
)

func (metadata DenomAuthorityMetadata) Validate() error {
	if bytes.Equal(metadata.Admin, []byte{}) {
		return errors.New("empty admin")
	}
	return nil
}
