package connection

import ()

type Connection interface {
	GetCounterparty() string
	GetClient() string
	GetCounterpartyClient() string

	Available() bool
}
