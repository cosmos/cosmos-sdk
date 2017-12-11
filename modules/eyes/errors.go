package eyes

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/errors"
)

var (
	errMissingData = fmt.Errorf("All tx fields must be filled")

	malformed = errors.CodeTypeEncodingErr
)

//nolint
func ErrMissingData() errors.TMError {
	return errors.WithCode(errMissingData, malformed)
}
func IsMissingDataErr(err error) bool {
	return errors.IsSameError(errMissingData, err)
}
