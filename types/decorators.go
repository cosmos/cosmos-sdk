package sdk

// A Decorator executes before/during/after a handler to enhance functionality.
type Decorator interface {

	// Decorate Handler.CheckTx
	CheckTx(ctx Context, ms MultiStore, tx Tx,
		next CheckTxFunc) CheckResult

	// Decorate Handler.DeliverTx
	DeliverTx(ctx Context, ms MultiStore, tx Tx,
		next DeliverTxFunc) DeliverResult
}

// A Decorator tied to its base handler "next" is itself a handler.
func Decorate(dec Decorator, next Handler) Handler {
	return &decHandler{
		decorator: dec,
		next:      next,
	}
}

//----------------------------------------

/*
	Helper to construct a decorated Handler from a stack of Decorators
	(first-decorator-first-call as in Python @decorators) , w/ Handler provided
	last for syntactic sugar of Stack().WithHandler()

	Usage:

	handler := sdk.Stack(
		decorator1,
		decorator2,
		...,
	).WithHandler(myHandler)

*/
func Stack(decorators ...Decorator) stack {
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
func (s *Stack) WithHandler(handler Handler) Handler {
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
	return decHandler{
		decorator: stack[0],
		next:      build(stack[1:], end),
	}
}

//----------------------------------------

type decHandler struct {
	decorator Decorator
	next      Handler
}

var _ Handler = &decHandler{}

func (dh *decHandler) CheckTx(ctx Context, ms MultiStore, tx Tx) CheckResult {
	return dh.decorator.CheckTx(ctx, ms, tx, dh.next)
}

func (dh *decHandler) DeliverTx(ctx Context, ms MultiStore, tx Tx) DeliverResult {
	return dh.decorator.DeliverTx(ctx, ms, tx, dh.next)
}
