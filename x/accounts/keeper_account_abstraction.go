package accounts

import (
	"context"
	"errors"
	"fmt"

	account_abstractionv1 "cosmossdk.io/x/accounts/interfaces/account_abstraction/v1"
	"cosmossdk.io/x/accounts/internal/implementation"
	v1 "cosmossdk.io/x/accounts/v1"
)

var (
	// ErrAuthentication is returned when the authentication fails.
	ErrAuthentication = errors.New("authentication failed")
	// ErrBundlerPayment is returned when the bundler payment fails.
	ErrBundlerPayment = errors.New("bundler payment failed")
	// ErrExecution is returned when the execution fails.
	ErrExecution = errors.New("execution failed")
)

// ExecuteUserOperation handles the execution of an abstracted account UserOperation.
func (k Keeper) ExecuteUserOperation(
	ctx context.Context,
	bundler string,
	op *v1.UserOperation,
) *v1.UserOperationResponse {
	resp := &v1.UserOperationResponse{}

	// authenticate
	authGas, err := k.Authenticate(ctx, bundler, op)
	resp.AuthenticationGasUsed = authGas
	if err != nil {
		resp.Error = err.Error()
		return resp
	}
	resp.AuthenticationGasUsed = authGas

	// pay bundler
	bundlerPayGas, bundlerPayResp, err := k.PayBundler(ctx, bundler, op)
	resp.BundlerPaymentGasUsed = bundlerPayGas
	if err != nil {
		resp.Error = err.Error()
		return resp
	}
	resp.BundlerPaymentResponses = bundlerPayResp

	// execute messages, the real operation intent
	executeGas, executeResp, err := k.OpExecuteMessages(ctx, bundler, op)
	resp.ExecutionGasUsed = executeGas
	if err != nil {
		resp.Error = err.Error()
		return resp
	}
	resp.ExecutionResponses = executeResp

	// done!
	return resp
}

// Authenticate handles the authentication flow of an abstracted account.
// Authentication happens in an isolated context with the authentication gas limit.
// If the authentication is successful, then the state is committed.
func (k Keeper) Authenticate(
	ctx context.Context,
	bundler string,
	op *v1.UserOperation,
) (gasUsed uint64, err error) {
	// authenticate
	gasUsed, err = k.branchExecutor.ExecuteWithGasLimit(ctx, op.AuthenticationGasLimit, func(ctx context.Context) error {
		return k.authenticate(ctx, bundler, op)
	})
	if err != nil {
		return gasUsed, fmt.Errorf("%v: %w", ErrAuthentication, err)
	}
	return gasUsed, nil
}

// authenticate handles the authentication flow of an abstracted account.
func (k Keeper) authenticate(
	ctx context.Context,
	bundler string,
	op *v1.UserOperation,
) error {
	senderAddr, err := k.addressCodec.StringToBytes(op.Sender)
	if err != nil {
		return err
	}
	// create an isolated context in which we execute authentication
	// without affecting the parent context and with the authentication gas limit.
	_, err = k.Execute(ctx, senderAddr, ModuleAccountAddress, &account_abstractionv1.MsgAuthenticate{
		Bundler:       bundler,
		UserOperation: op,
	})
	return err
}

// OpExecuteMessages handles the execution of the messages in a given v1.UserOperation.
// It executes in an isolated branch, in an atomic way, if all the messages pass then
// the execution is deemed successful and the state is committed.
// An account abstraction implementer can choose to handle execution messages or not,
// if it does not expose the execution messages method, then this method will simply
// execute the provided messages on behalf of the sender and return.
func (k Keeper) OpExecuteMessages(
	ctx context.Context,
	bundler string,
	op *v1.UserOperation,
) (gasUsed uint64, responses []*implementation.Any, err error) {
	// execute messages, the real operation intent
	gasUsed, err = k.branchExecutor.ExecuteWithGasLimit(ctx, op.ExecutionGasLimit, func(ctx context.Context) error {
		responses, err = k.opExecuteMessages(ctx, bundler, op)
		return err
	})
	if err != nil {
		return gasUsed, nil, fmt.Errorf("%v: %w", ErrExecution, err)
	}
	return gasUsed, responses, nil
}

func (k Keeper) opExecuteMessages(
	ctx context.Context,
	bundler string,
	op *v1.UserOperation,
) (messagesResponse []*implementation.Any, err error) {
	senderAddr, err := k.addressCodec.StringToBytes(op.Sender)
	if err != nil {
		return nil, err
	}
	resp, err := k.Execute(ctx, senderAddr, ModuleAccountAddress, &account_abstractionv1.MsgExecute{
		Bundler:           bundler,
		ExecutionMessages: op.ExecutionMessages,
	})
	// here is where we check if the account handles execution messages
	// if it does not, then we simply execute the provided messages on behalf of the sender
	switch {
	case err == nil:
		// all is ok, so parse responses.
		executeResp, err := parseExecuteResponse(resp)
		return executeResp, err
	case implementation.IsRoutingError(err):
		// if it is a routing error, it means the account does not handle execution messages,
		// in this case we attempt to execute the provided messages on behalf of the op sender.
		return k.sendAnyMessages(ctx, senderAddr, op.ExecutionMessages)
	default:
		// some other error
		return nil, err
	}
}

// PayBundler handles the payment of the bundler in a given v1.UserOperation.
// Must be called after Authenticate.
// It gets executed in an isolated context with the bundler payment gas limit.
// If the payment is successful, then the state is committed.
// Since for an abstracted account the bundler payment method is optional,
// if the account does not handle bundler payment messages, then this method
// will simply execute the provided messages on behalf of the sender and return.
func (k Keeper) PayBundler(
	ctx context.Context,
	bundler string,
	op *v1.UserOperation,
) (gasUsed uint64, responses []*implementation.Any, err error) {
	// pay bundler
	gasUsed, err = k.branchExecutor.ExecuteWithGasLimit(ctx, op.BundlerPaymentGasLimit, func(ctx context.Context) error {
		responses, err = k.payBundler(ctx, bundler, op)
		return err
	})
	if err != nil {
		return gasUsed, nil, fmt.Errorf("%v: %w", ErrBundlerPayment, err)
	}
	return gasUsed, responses, nil
}

func (k Keeper) payBundler(
	ctx context.Context,
	bundler string,
	op *v1.UserOperation,
) (paymentResponses []*implementation.Any, err error) {
	// if messages are empty, then there is nothing to do
	if len(op.BundlerPaymentMessages) == 0 {
		return nil, nil
	}
	// pay bundler
	senderAddr, err := k.addressCodec.StringToBytes(op.Sender)
	if err != nil {
		return nil, err
	}
	resp, err := k.Execute(ctx, senderAddr, ModuleAccountAddress, &account_abstractionv1.MsgPayBundler{
		Bundler:                bundler,
		BundlerPaymentMessages: op.BundlerPaymentMessages,
	})
	// here is where we check if the account handles bundler payment messages
	// if it does not, then we simply execute the provided messages on behalf of the sender
	switch {
	case err == nil:
		// if no error, execution went fine, so parse responses.
		payBundlerResp, err := parsePayBundlerResponse(resp)
		return payBundlerResp, err
	case implementation.IsRoutingError(err):
		// if we get a routing message error it means the account does not handle bundler payment messages,
		// in this case we attempt to execute the provided messages on behalf of the op sender.
		return k.sendAnyMessages(ctx, senderAddr, op.BundlerPaymentMessages)
	default:
		// some other execution error.
		return nil, err
	}
}

// parsePayBundlerResponse parses the bundler response as any into a slice of
// responses on payment messages.
func parsePayBundlerResponse(resp any) ([]*implementation.Any, error) {
	payBundlerResp, ok := resp.(*account_abstractionv1.MsgPayBundlerResponse)
	// this means the account does not properly implement account abstraction.
	if payBundlerResp == nil {
		return nil, fmt.Errorf("account does not implement account abstraction correctly: wanted %T, got nil", &account_abstractionv1.MsgPayBundlerResponse{})
	}
	if !ok {
		return nil, fmt.Errorf("account does not implement account abstraction correctly: wanted %T, got %T", &account_abstractionv1.MsgPayBundlerResponse{}, resp)
	}
	return payBundlerResp.BundlerPaymentMessagesResponse, nil
}

// parseExecuteResponse parses the execute response as any into a slice of
// responses on execution messages.
func parseExecuteResponse(resp any) ([]*implementation.Any, error) {
	executeResp, ok := resp.(*account_abstractionv1.MsgExecuteResponse)
	// this means the account does not properly implement account abstraction.
	if executeResp == nil {
		return nil, fmt.Errorf("account does not implement account abstraction correctly: wanted %T, got nil", &account_abstractionv1.MsgExecuteResponse{})
	}
	if !ok {
		return nil, fmt.Errorf("account does not implement account abstraction correctly: wanted %T, got %T", &account_abstractionv1.MsgExecuteResponse{}, resp)
	}
	return executeResp.ExecutionMessagesResponse, nil
}
