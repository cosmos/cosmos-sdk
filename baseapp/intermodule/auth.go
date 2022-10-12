package intermodule

import "context"

type Authorizer func(ctx context.Context, methodName string, req interface{}, callingModule string) bool
