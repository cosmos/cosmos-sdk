package common

import (
	"time"
)

type (
	// Registry is an abstract entry point to register message factories with weights
	Registry interface {
		Add(weight uint32, f SimMsgFactoryX)
	}
	// FutureOpsRegistry register message factories for future blocks
	FutureOpsRegistry interface {
		Add(blockTime time.Time, f SimMsgFactoryX)
	}

	HasFutureOpsRegistry interface {
		SetFutureOpsRegistry(FutureOpsRegistry)
	}
)
