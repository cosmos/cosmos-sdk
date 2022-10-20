package appmodule

import (
	"context"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type UpgradeRegistrar interface {
	RegisterUpgradeHandler(fromModule protoreflect.FullName, handler func(context.Context) error)
}
