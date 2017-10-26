package sdk

// Decorator is anything that wraps another handler
// to enhance functionality.
//
// They are usually chained together via ChainDecorators
// before wrapping an interface.
type Decorator interface {
	DecorateChecker
	DecorateDeliverer
}

type DecorateChecker interface {
	CheckTx(ctx Context, store SimpleDB,
		tx interface{}, next Checker) (CheckResult, error)
}

type DecorateDeliverer interface {
	DeliverTx(ctx Context, store SimpleDB, tx interface{},
		next Deliverer) (DeliverResult, error)
}

// Stack is the entire application stack
type Stack struct {
	decorators []Decorator
	handler    Handler
	Handler    // the compiled version, which we expose
}

var _ Handler = &Stack{}

// ChainDecorators prepares a stack of decorators,
// you must call `.WithHandler()` before you can execute it.
func ChainDecorators(decorators ...Decorator) *Stack {
	s := &Stack{
		decorators: decorators,
	}
	return s
}

// WithHandler sets the final handler for the stack and
// prepares it for use
func (s *Stack) WithHandler(handler Handler) *Stack {
	if handler == nil {
		panic("Cannot have a Stack without an end handler")
	}
	s.handler = handler
	s.Handler = build(s.decorators, s.handler)
	return s
}

// build wraps each decorator around the next, so that
// the last in the list is closest to the handler
func build(stack []Decorator, end Handler) Handler {
	if len(stack) == 0 {
		return end
	}
	return wrap(stack[0], build(stack[1:], end))
}

// decorator lets us wrap a whole stack up into one Handler
//
// heavily inspired by negroni's design
type decorator struct {
	decorator Decorator
	next      Handler
}

// ensure it fulfils the interface
var _ Handler = &decorator{}

// CheckTx fulfils Handler interface
func (m *decorator) CheckTx(ctx Context, store SimpleDB,
	tx interface{}) (CheckResult, error) {

	return m.decorator.CheckTx(ctx, store, tx, m.next)
}

// DeliverTx fulfils Handler interface
func (m *decorator) DeliverTx(ctx Context, store SimpleDB,
	tx interface{}) (res DeliverResult, err error) {

	return m.decorator.DeliverTx(ctx, store, tx, m.next)
}

// wrap puts one decorator around a handler
func wrap(dec Decorator, next Handler) Handler {
	return &decorator{
		decorator: dec,
		next:      next,
	}
}
