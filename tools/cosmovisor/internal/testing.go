package internal

import "context"

type TestCallback func()

type testCallbackKey struct{}

func WithTestCallback(ctx context.Context, cb TestCallback) context.Context {
	return context.WithValue(ctx, testCallbackKey{}, cb)
}

func GetTestCallback(ctx context.Context) TestCallback {
	cb, ok := ctx.Value(testCallbackKey{}).(TestCallback)
	if !ok {
		return nil
	}
	return cb
}
