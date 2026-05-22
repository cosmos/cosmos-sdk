package types

import "context"

// strictWithdrawCtxKey marks a context as originating from a user tx, signaling
// to downstream hooks that errors should propagate instead of falling back.
type strictWithdrawCtxKey struct{}

// WithStrictWithdraw marks ctx so hooks propagate errors instead of falling back.
func WithStrictWithdraw(ctx context.Context) context.Context {
	return context.WithValue(ctx, strictWithdrawCtxKey{}, true)
}

// IsStrictWithdraw reports whether ctx was marked by WithStrictWithdraw.
func IsStrictWithdraw(ctx context.Context) bool {
	v, _ := ctx.Value(strictWithdrawCtxKey{}).(bool)
	return v
}
