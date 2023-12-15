package appmanager

import (
	"context"
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/serverv2/core/appmanager"
)

func NewSTFBuilder() *STFBuilder {
	return &STFBuilder{
		err:                nil,
		msgRouterBuilder:   newMsgRouterBuilder(),
		queryRouterBuilder: newMsgRouterBuilder(),
		txValidators:       make(map[string]func(ctx context.Context, tx Tx) error),
		beginBlockers:      make(map[string]func(ctx context.Context) error),
		endBlockers:        make(map[string]func(ctx context.Context) error),
	}
}

type STFBuilder struct {
	err error

	msgRouterBuilder   *msgRouterBuilder
	queryRouterBuilder *msgRouterBuilder
	txValidators       map[string]func(ctx context.Context, tx Tx) error
	beginBlockers      map[string]func(ctx context.Context) error
	endBlockers        map[string]func(ctx context.Context) error
	txCodec            TxDecoder
}

type STFBuilderOptions struct {
	// OrderEndBlockers can be optionally provided to set the order of end blockers.
	OrderEndBlockers []string
	// OrderBeginBlockers can be optionally provided to set the order of begin blockers.
	OrderBeginBlockers []string
	// OrderTxValidators can be optionally provided to set the order of tx validators.
	OrderTxValidators []string
}

func (s *STFBuilder) Build(opts *STFBuilderOptions) (*STFAppManager, error) {
	msgHandler, err := s.msgRouterBuilder.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to build msg handler: %w", err)
	}
	queryHandler, err := s.queryRouterBuilder.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to build query handler: %w", err)
	}
	beginBlocker, err := s.makeBeginBlocker(opts.OrderBeginBlockers)
	if err != nil {
		return nil, fmt.Errorf("unable to build begin blocker: %w", err)
	}
	endBlocker, err := s.makeEndBlocker(opts.OrderEndBlockers)
	if err != nil {
		return nil, fmt.Errorf("unable to build end blocker: %w", err)
	}
	txValidator, err := s.makeTxValidator(opts.OrderTxValidators)
	if err != nil {
		return nil, fmt.Errorf("unable to build tx validator: %w", err)
	}
	return &STFAppManager{
		handleMsg:      msgHandler,
		handleQuery:    queryHandler,
		doBeginBlock:   beginBlocker,
		doEndBlock:     endBlocker,
		doTxValidation: txValidator,
		decodeTx: func(txBytes []byte) (Tx, error) {
			return s.txCodec.Decode(txBytes)
		},
		branch: nil, // TODO
	}, nil
}

func (s *STFBuilder) AddModule(m appmanager.Module) {
	// TODO: the best is add modules but not build them here but build them later when we call STFBuilder.Build.
	// build msg handler
	moduleMsgRouter := _newModuleMsgRouter(m.Name(), s.msgRouterBuilder)
	m.RegisterMsgHandlers(moduleMsgRouter)
	m.RegisterPreMsgHandler(moduleMsgRouter)
	m.RegisterPostMsgHandler(moduleMsgRouter)
	// build query handler
	moduleQueryRouter := _newModuleMsgRouter(m.Name(), s.queryRouterBuilder)
	m.RegisterQueryHandler(moduleQueryRouter)
	// add begin blockers and endblockers
	// TODO: check if is not nil, etc.
	s.beginBlockers[m.Name()] = m.BeginBlocker()
	s.endBlockers[m.Name()] = m.EndBlocker()
	s.txValidators[m.Name()] = m.TxValidator()
}

func (s *STFBuilder) makeEndBlocker(order []string) (func(ctx context.Context) error, error) {
	// TODO do ordering...
	// TODO do checks if all are present etc
	return func(ctx context.Context) error {
		for module, f := range s.endBlockers {
			err := f(ctx)
			if err != nil {
				return fmt.Errorf("endblocker of module %s failure: %w", module, err)
			}
		}
		return nil
	}, nil
}

func (s *STFBuilder) makeBeginBlocker(order []string) (func(ctx context.Context) error, error) {
	// TODO do ordering...
	// TODO do checks if all are present etc
	return func(ctx context.Context) error {
		for module, f := range s.beginBlockers {
			err := f(ctx)
			if err != nil {
				return fmt.Errorf("beginblocker of module %s failure: %w", module, err)
			}
		}
		return nil
	}, nil
}

func (s *STFBuilder) makeTxValidator(order []string) (func(ctx context.Context, tx Tx) error, error) {
	// TODO do ordering...
	// TODO do checks if all are present etc
	return func(ctx context.Context, tx Tx) error {
		for module, f := range s.txValidators {
			err := f(ctx, tx)
			if err != nil {
				return fmt.Errorf("tx validation failed for module %s: %w", module, err)
			}
		}
		return nil
	}, nil
}

// we create some intermediary type that associates a registration error with the module.
var _ appmanager.MsgRouterBuilder = (*_moduleMsgRouter)(nil)
var _ appmanager.PostMsgRouterBuilder = (*_moduleMsgRouter)(nil)
var _ appmanager.PreMsgRouterBuilder = (*_moduleMsgRouter)(nil)

func _newModuleMsgRouter(moduleName string, router *msgRouterBuilder) *_moduleMsgRouter {
	return &_moduleMsgRouter{
		err:              nil,
		moduleName:       moduleName,
		msgRouterBuilder: router,
	}
}

type _moduleMsgRouter struct {
	err              error
	moduleName       string
	msgRouterBuilder *msgRouterBuilder
}

func (r *_moduleMsgRouter) RegisterPreHandler(msg appmanager.Type, preHandler func(ctx context.Context, msg appmanager.Type) error) {
	r.msgRouterBuilder.RegisterPreHandler(TypeName(msg), preHandler)
}

func (r *_moduleMsgRouter) RegisterPostHandler(msg appmanager.Type, postHandler func(ctx context.Context, msg appmanager.Type, msgResp appmanager.Type) error) {
	r.msgRouterBuilder.RegisterPostHandler(TypeName(msg), postHandler)
}

func (r *_moduleMsgRouter) RegisterHandler(msg appmanager.Type, handlerFunc func(ctx context.Context, msg appmanager.Type) (resp appmanager.Type, err error)) {
	err := r.msgRouterBuilder.RegisterHandler(TypeName(msg), handlerFunc)
	if err != nil {
		r.err = errors.Join(r.err, fmt.Errorf("%w: %s", err, r.moduleName))
	}
}
