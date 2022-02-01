package ormerrors

import "github.com/cosmos/cosmos-sdk/errors"

var codespace = "orm"

var (
	InvalidTableId                = errors.New(codespace, 1, "invalid or missing table or single id, need a non-zero value")
	MissingPrimaryKey             = errors.New(codespace, 2, "table is missing primary key")
	InvalidKeyFieldsDefinition    = errors.New(codespace, 3, "invalid field definition for key")
	DuplicateKeyField             = errors.New(codespace, 4, "duplicate field in key")
	FieldNotFound                 = errors.New(codespace, 5, "field not found")
	InvalidAutoIncrementKey       = errors.New(codespace, 6, "an auto-increment primary key must specify a single uint64 field")
	InvalidIndexId                = errors.New(codespace, 7, "invalid or missing index id, need a value >= 0 and < 32768")
	DuplicateIndexId              = errors.New(codespace, 8, "duplicate index id")
	PrimaryKeyConstraintViolation = errors.New(codespace, 9, "object with primary key already exists")
	NotFoundOnUpdate              = errors.New(codespace, 10, "can't update object which doesn't exist")
	PrimaryKeyInvalidOnUpdate     = errors.New(codespace, 11, "can't update object with missing or invalid primary key")
	AutoIncrementKeyAlreadySet    = errors.New(codespace, 12, "can't create with auto-increment primary key already set")
	CantFindIndex                 = errors.New(codespace, 13, "can't find index")
	UnexpectedDecodePrefix        = errors.New(codespace, 14, "unexpected prefix while trying to decode an entry")
	BytesFieldTooLong             = errors.New(codespace, 15, "bytes field is longer than 255 bytes")
	UnsupportedOperation          = errors.New(codespace, 16, "unsupported operation")
	BadDecodeEntry                = errors.New(codespace, 17, "bad decode entry")
	IndexOutOfBounds              = errors.New(codespace, 18, "index out of bounds")
	InvalidListOptions            = errors.New(codespace, 19, "invalid list options")
	UnsupportedKeyField           = errors.New(codespace, 20, "unsupported key field")
	UnexpectedError               = errors.New(codespace, 21, "unexpected error")
	InvalidRangeIterationKeys     = errors.New(codespace, 22, "invalid range iteration keys")
	JSONImportError               = errors.New(codespace, 23, "json import error")
	UniqueKeyViolation            = errors.New(codespace, 24, "unique key violation")
	InvalidTableDefinition        = errors.New(codespace, 25, "invalid table definition")
	InvalidFileDescriptorID       = errors.New(codespace, 26, "invalid file descriptor ID")
	TableNotFound                 = errors.New(codespace, 27, "table not found")
)
