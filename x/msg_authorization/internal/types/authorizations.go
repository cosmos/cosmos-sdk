package types

import (
	"time"

	"github.com/cosmos/cosmos-sdk/x/msg_authorization/exported"
)

type AuthorizationGrant struct {
	Authorization exported.Authorization

	Expiration time.Time
}
