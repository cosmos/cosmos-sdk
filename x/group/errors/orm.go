package errors

import sdkerrors "cosmossdk.io/errors"

// mathCodespace is the codespace for all errors defined in orm package
const ormCodespace = "orm"

var (
	// ErrORMIteratorDone defines an error when an iterator is done
	ErrORMIteratorDone = sdkerrors.Register(ormCodespace, 11, "iterator done")

	// ErrORMInvalidIterator defines an error for an invalid iterator
	ErrORMInvalidIterator = sdkerrors.Register(ormCodespace, 12, "invalid iterator")

	// ErrORMUniqueConstraint defines an error when a value already exists at a given key
	ErrORMUniqueConstraint = sdkerrors.Register(ormCodespace, 13, "unique constraint violation")

	// ErrORMInvalidArgument defines an error when an invalid argument is provided as part of ORM functions
	ErrORMInvalidArgument = sdkerrors.Register(ormCodespace, 14, "invalid argument")

	// ErrORMKeyMaxLength defines an error when a key exceeds max length
	ErrORMKeyMaxLength = sdkerrors.Register(ormCodespace, 15, "key exceeds max length")

	// ErrORMEmptyKey defines an error for an empty key
	ErrORMEmptyKey = sdkerrors.Register(ormCodespace, 47, "cannot use empty key")
)
