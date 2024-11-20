package types

import (
	"errors"
)

func (metadata DenomAuthorityMetadata) Validate() error {
	if metadata.Admin == "" {
		return errors.New("empty admin")
	}
	return nil
}
