package ormerrors

import (
	"cosmossdk.io/errors"
	"google.golang.org/grpc/codes"
)

var codespace = "orm"

// IsErrNotFound returns true if the error indicates that the record was not found.
func IsErrNotFound(err error) bool {
	return errors.IsOf(err, ErrNotFound)
}

var (
	ErrInvalidTableId                   = errors.New(codespace, 1, "invalid or missing table or single id, need a non-zero value")
	ErrMissingPrimaryKey                = errors.New(codespace, 2, "table is missing primary key")
	ErrErrInvalidKeyFieldsDefinition    = errors.New(codespace, 3, "invalid field definition for key")
	ErrIDuplicateKeyField               = errors.New(codespace, 4, "duplicate field in key")
	ErrFieldErrNotFound                 = errors.New(codespace, 5, "field not found")
	ErrInvalidAutoIncrementKey          = errors.New(codespace, 6, "an auto-increment primary key must specify a single uint64 field")
	ErrInvalidIndexId                   = errors.New(codespace, 7, "invalid or missing index id, need a value >= 0 and < 32768")
	ErrDuplicateIndexId                 = errors.New(codespace, 8, "duplicate index id")
	ErrPrimaryKeyErrConstraintViolation = errors.New(codespace, 9, "object with primary key already exists")
	ErrPrimaryKeyInvalidOnUpdate        = errors.New(codespace, 11, "can't update object with missing or invalid primary key")
	ErrAutoIncrementKeyAlreadySet       = errors.New(codespace, 12, "can't create with auto-increment primary key already set")
	ErrCantFindIndex                    = errors.New(codespace, 13, "can't find index")
	ErrUnexpectedDecodePrefix           = errors.New(codespace, 14, "unexpected prefix while trying to decode an entry")
	ErrUnsupportedOperation             = errors.New(codespace, 16, "unsupported operation")
	ErrBadDecodeEntry                   = errors.New(codespace, 17, "bad decode entry")
	ErrIndexOutOfBounds                 = errors.New(codespace, 18, "index out of bounds")
	ErrInvalidListOptions               = errors.New(codespace, 19, "invalid list options")
	ErrInvalidKeyField                  = errors.New(codespace, 20, "invalid key field")
	ErrUnexpectedError                  = errors.New(codespace, 21, "unexpected error")
	ErrInvalidRangeIterationKeys        = errors.New(codespace, 22, "invalid range iteration keys")
	ErrJSONImportError                  = errors.New(codespace, 23, "json import error")
	ErrUniqueKeyViolation               = errors.RegisterWithGRPCCode(codespace, 24, codes.FailedPrecondition, "unique key violation")
	ErrInvalidTableDefinition           = errors.New(codespace, 25, "invalid table definition")
	ErrInvalidFileDescriptorID          = errors.New(codespace, 26, "invalid file descriptor ID")
	ErrTableErrNotFound                 = errors.New(codespace, 27, "table not found")
	ErrJSONValidationError              = errors.New(codespace, 28, "invalid JSON")
	ErrNotFound                         = errors.RegisterWithGRPCCode(codespace, 29, codes.NotFound, "not found")
	ErrReadOnly                         = errors.New(codespace, 30, "database is read-only")
	ErrAlreadyExists                    = errors.RegisterWithGRPCCode(codespace, 31, codes.AlreadyExists, "already exists")
	ErrConstraintViolation              = errors.RegisterWithGRPCCode(codespace, 32, codes.FailedPrecondition, "failed precondition")
	ErrNoTableDescriptor                = errors.New(codespace, 33, "no table descriptor found")
)
