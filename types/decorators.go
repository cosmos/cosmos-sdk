package types

// A Decorator executes before/during/after a handler to enhance functionality.
type Decorator func(ctx Context, ms MultiStore, tx Tx, next Handler) Result

// Return a decorated handler
func Decorate(dec Decorator, next Handler) Handler {
	return func(ctx Context, ms MultiStore, tx Tx) Result {
		return dec(ctx, ms, tx, next)
	}
}

//----------------------------------------

/*
	Helper to construct a decorated Handler from a stack of Decorators
	(first-decorator-first-call as in Python @decorators) , w/ Handler provided
	last for syntactic sugar of ChainDecorators().WithHandler()

	Usage:

	handler := sdk.ChainDecorators(
		decorator1,
		decorator2,
		...,
	).WithHandler(myHandler)

*/
func ChainDecorators(decorators ...Decorator) stack {
	return stack{
		decs: decorators,
	}
}

// No need to expose this.
type stack struct {
	decs []Decorator
}

// WithHandler sets the final handler for the stack and
// returns the decoratored Handler.
func (s stack) WithHandler(handler Handler) Handler {
	if handler == nil {
		panic("WithHandler() requires a non-nil Handler")
	}
	return build(s.decs, handler)
}

// build wraps each decorator around the next, so that
// the last in the list is closest to the handler
func build(stack []Decorator, end Handler) Handler {
	if len(stack) == 0 {
		return end
	}
	return Decorate(stack[0], build(stack[1:], end))
}
