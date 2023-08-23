package implementation

import "context"

// NewImplementation creates a new Implementation instance given an Account implementer.
func NewImplementation(account Account) (Implementation, error) {
	// make init handler
	ir := NewInitBuilder()
	account.RegisterInitHandler(ir)
	initHandler, err := ir.makeHandler()
	if err != nil {
		return Implementation{}, err
	}

	// make execute handler
	er := NewExecuteBuilder()
	account.RegisterExecuteHandlers(er)
	executeHandler, err := er.makeHandler()
	if err != nil {
		return Implementation{}, err
	}

	// make query handler
	qr := NewQueryBuilder()
	account.RegisterQueryHandlers(qr)
	queryHandler, err := qr.makeHandler()
	if err != nil {
		return Implementation{}, err
	}
	return Implementation{
		Init:    initHandler,
		Execute: executeHandler,
		Query:   queryHandler,
	}, nil
}

// Implementation wraps an Account implementer in order to provide a concrete
// and non-generic implementation usable by the x/accounts module.
type Implementation struct {
	// Init defines the initialisation handler for the smart account.
	Init func(ctx context.Context, msg any) (resp any, err error)
	// Execute defines the execution handler for the smart account.
	Execute func(ctx context.Context, msg any) (resp any, err error)
	// Query defines the query handler for the smart account.
	Query func(ctx context.Context, msg any) (resp any, err error)
}
