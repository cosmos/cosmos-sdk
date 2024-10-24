package types

import (
	"errors"
)

var (
	// ErrNoAuthorizationFound error if there is no authorization found given a grantee address
	ErrNoAuthorizationFound = errors.New("authorization not found")
	ErrInvalidAuthorization = errors.New("invalid authorization")
	// ErrInvalidExpirationTime error if the set expiration time is in the past
	ErrInvalidExpirationTime = errors.New("expiration time of authorization should be more than current time")
	// ErrUnknownAuthorizationType error for unknown authorization type
	ErrUnknownAuthorizationType = errors.New("unknown authorization type")
	// ErrAuthorizationExpired error if the authorization has expired
	ErrAuthorizationExpired = errors.New("authorization expired")
	// ErrGranteeIsGranter error if the grantee and the granter are the same
	ErrGranteeIsGranter = errors.New("grantee and granter should be different")
	// ErrUnauthorizedAction error if the sender address not the account granter address
	ErrUnauthorizedAction = errors.New("unauthorized action, only account granter can perform this action")
)
