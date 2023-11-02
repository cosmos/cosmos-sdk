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
		Init:                  initHandler,
		Execute:               executeHandler,
		Query:                 queryHandler,
		DecodeInitRequest:     ir.decodeRequest,
		EncodeInitResponse:    ir.encodeResponse,
		DecodeExecuteRequest:  er.makeRequestDecoder(),
		EncodeExecuteResponse: er.makeResponseEncoder(),
		DecodeQueryRequest:    qr.er.makeRequestDecoder(),
		EncodeQueryResponse:   qr.er.makeResponseEncoder(),
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

	// Schema

	// DecodeInitRequest decodes an init request coming from the message server.
	DecodeInitRequest func([]byte) (any, error)
	// EncodeInitResponse encodes an init response to be sent back from the message server.
	EncodeInitResponse func(any) ([]byte, error)

	// DecodeExecuteRequest decodes an execute request coming from the message server.
	DecodeExecuteRequest func([]byte) (any, error)
	// EncodeExecuteResponse encodes an execute response to be sent back from the message server.
	EncodeExecuteResponse func(any) ([]byte, error)

	// DecodeQueryRequest decodes a query request coming from the message server.
	DecodeQueryRequest func([]byte) (any, error)
	// EncodeQueryResponse encodes a query response to be sent back from the message server.
	EncodeQueryResponse func(any) ([]byte, error)
}
