package errors

import (
	errorsmod "cosmossdk.io/errors"
)

// mathCodespace is the codespace for all errors defined in orm package
const ormCodespace = "orm"

var (
	// ErrORMIteratorDone defines an error when an iterator is done
	ErrORMIteratorDone = errorsmod.Register(ormCodespace, 11, "iterator done")

	// ErrORMInvalidIterator defines an error for an invalid iterator
	ErrORMInvalidIterator = errorsmod.Register(ormCodespace, 12, "invalid iterator")

	// ErrORMUniqueConstraint defines an error when a value already exists at a given key
	ErrORMUniqueConstraint = errorsmod.Register(ormCodespace, 13, "unique constraint violation")

	// ErrORMInvalidArgument defines an error when an invalid argument is provided as part of ORM functions
	ErrORMInvalidArgument = errorsmod.Register(ormCodespace, 14, "invalid argument")

	// ErrORMKeyMaxLength defines an error when a key exceeds max length
	ErrORMKeyMaxLength = errorsmod.Register(ormCodespace, 15, "key exceeds max length")

	// ErrORMEmptyKey defines an error for an empty key
	ErrORMEmptyKey = errorsmod.Register(ormCodespace, 47, "cannot use empty key")
)
