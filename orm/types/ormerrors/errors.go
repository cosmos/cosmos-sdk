package ormerrors

import (
	"cosmossdk.io/errors"
	"google.golang.org/grpc/codes"
)

var ormcodespace = "orm"

// IsNotFound returns true if the error indicates that the record was not found.
func IsNotFound(err error) bool {
	return errors.IsOf(err, ErrNotFound)
}

var (
	ErrInvalidTableID                = errors.New(ormcodespace, 1, "invalid or missing table or single id, need a non-zero value")
	ErrMissingPrimaryKey             = errors.New(ormcodespace, 2, "table is missing primary key")
	ErrInvalidKeyFieldsDefinition    = errors.New(ormcodespace, 3, "invalid field definition for key")
	ErrDuplicateKeyField             = errors.New(ormcodespace, 4, "duplicate field in key")
	ErrFieldNotFound                 = errors.New(ormcodespace, 5, "field not found")
	ErrInvalidAutoIncrementKey       = errors.New(ormcodespace, 6, "an auto-increment primary key must specify a single uint64 field")
	ErrInvalidIndexID                = errors.New(ormcodespace, 7, "invalid or missing index id, need a value >= 0 and < 32768")
	ErrDuplicateIndexID              = errors.New(ormcodespace, 8, "duplicate index id")
	ErrPrimaryKeyConstraintViolation = errors.New(ormcodespace, 9, "object with primary key already exists")
	ErrPrimaryKeyInvalidOnUpdate     = errors.New(ormcodespace, 11, "can't update object with missing or invalid primary key")
	ErrAutoIncrementKeyAlreadySet    = errors.New(ormcodespace, 12, "can't create with auto-increment primary key already set")
	ErrCantFindIndex                 = errors.New(ormcodespace, 13, "can't find index")
	ErrUnexpectedDecodePrefix        = errors.New(ormcodespace, 14, "unexpected prefix while trying to decode an entry")
	ErrUnsupportedOperation          = errors.New(ormcodespace, 16, "unsupported operation")
	ErrBadDecodeEntry                = errors.New(ormcodespace, 17, "bad decode entry")
	ErrIndexOutOfBounds              = errors.New(ormcodespace, 18, "index out of bounds")
	ErrInvalidListOptions            = errors.New(ormcodespace, 19, "invalid list options")
	ErrInvalidKeyField               = errors.New(ormcodespace, 20, "invalid key field")
	ErrUnexpectedError               = errors.New(ormcodespace, 21, "unexpected error")
	ErrInvalidRangeIterationKeys     = errors.New(ormcodespace, 22, "invalid range iteration keys")
	ErrJSONImportError               = errors.New(ormcodespace, 23, "json import error")
	ErrUniqueKeyViolation            = errors.RegisterWithGRPCCode(ormcodespace, 24, codes.FailedPrecondition, "unique key violation")
	ErrInvalidTableDefinition        = errors.New(ormcodespace, 25, "invalid table definition")
	ErrInvalidFileDescriptorID       = errors.New(ormcodespace, 26, "invalid file descriptor ID")
	ErrTableNotFound                 = errors.New(ormcodespace, 27, "table not found")
	ErrJSONValidationError           = errors.New(ormcodespace, 28, "invalid JSON")
	ErrNotFound                      = errors.RegisterWithGRPCCode(ormcodespace, 29, codes.NotFound, "not found")
	ErrReadOnly                      = errors.New(ormcodespace, 30, "database is read-only")
	ErrOrmAlreadyExists              = errors.RegisterWithGRPCCode(ormcodespace, 31, codes.AlreadyExists, "already exists")
	ErrConstraintViolation           = errors.RegisterWithGRPCCode(ormcodespace, 32, codes.FailedPrecondition, "failed precondition")
	ErrNoTableDescriptor             = errors.New(ormcodespace, 33, "no table descriptor found")
)
