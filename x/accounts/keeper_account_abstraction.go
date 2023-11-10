package accounts

import (
	"bytes"
	"context"
	"fmt"

	account_abstractionv1 "cosmossdk.io/api/cosmos/accounts/interfaces/account_abstraction/v1"
	accountsv1 "cosmossdk.io/api/cosmos/accounts/v1"
	"cosmossdk.io/x/accounts/internal/implementation"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/runtime/protoiface"
	"google.golang.org/protobuf/types/known/anypb"
)

// ExecuteUserOperation handles the execution of an abstracted account UserOperation.
func (k Keeper) ExecuteUserOperation(ctx context.Context, bundler string, op *accountsv1.UserOperation) *accountsv1.UserOperationResponse {
	resp := &accountsv1.UserOperationResponse{}

	// authenticate
	authGas, err := k.Authenticate(ctx, bundler, op)
	if err != nil {
		resp.Error = fmt.Sprintf("authentication failed: %s", err.Error())
		return resp
	}
	resp.AuthenticationGasUsed = authGas

	// pay bundler
	bundlerPayGas, bundlerPayResp, err := k.PayBundler(ctx, bundler, op)
	if err != nil {
		resp.Error = fmt.Sprintf("bundler payment failed: %s", err.Error())
		return resp
	}
	resp.BundlerPaymentGasUsed = bundlerPayGas
	resp.BundlerPaymentResponses = bundlerPayResp

	// execute messages, the real operation intent
	executeGas, executeResp, err := k.ExecuteMessages(ctx, bundler, op)
	if err != nil {
		resp.Error = fmt.Sprintf("execution failed: %s", err.Error())
		return resp
	}
	resp.ExecutionGasUsed = executeGas
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
	op *accountsv1.UserOperation,
) (gasUsed uint64, err error) {
	// TODO: add branch with gas limit
	senderAddr, err := k.addressCodec.StringToBytes(op.Sender)
	if err != nil {
		return 0, err
	}
	accountNumber, err := k.AccountByNumber.Get(ctx, senderAddr)
	// create an isolated context in which we execute authentication
	// without affecting the parent context and with the authentication gas limit.
	_, err = k.Execute(ctx, senderAddr, ModuleAccountAddr, &account_abstractionv1.MsgAuthenticate{
		Bundler:       bundler,
		UserOperation: op,
		ChainId:       "chain-id", // TODO how to get chain id?
		AccountNumber: accountNumber,
	})
	return gasUsed, nil
}

// PayBundler handles the payment of the bundler in a given v1.UserOperation.
// Must be called after Authenticate.
// It gets executed in an isolated context with the bundler payment gas limit.
// If the payment is successful, then the state is committed.
// Since for an abstracted account the bundler payment method is optional,
// if the account does not handle bundler payment messages, then this method
// will simply execute the provided messages on behalf of the sender and return.
func (k Keeper) PayBundler(ctx context.Context, bundler string, op *accountsv1.UserOperation) (gasUsed uint64, paymentResponses []*anypb.Any, err error) {
	// TODO add branch with gas limit
	senderAddr, err := k.addressCodec.StringToBytes(op.Sender)
	if err != nil {
		return 0, nil, err
	}
	resp, err := k.Execute(ctx, senderAddr, ModuleAccountAddr, &account_abstractionv1.MsgPayBundler{
		Bundler:                bundler,
		BundlerPaymentMessages: op.BundlerPaymentMessages,
	})
	// here is where we check if the account handles bundler payment messages
	// if it does not, then we simply execute the provided messages on behalf of the sender
	switch {
	case err == nil:
		// TODO: get gas used
		payBundlerResp, err := parsePayBundlerResponse(resp)
		return gasUsed, payBundlerResp, err
	// if we get a routing message error it means the account does not handle bundler payment messages,
	// in this case we attempt to execute the provided messages on behalf of the op sender.
	case implementation.IsRoutingError(err):
		// TODO: get gas used
		payBundlerResp, err := k.payBundlerFallback(ctx, op)
		return gasUsed, payBundlerResp, err
	// some other execution error.
	default:
		// TODO: get gas used
		return gasUsed, nil, err
	}
}

// payBundlerFallback attempts to execute the provided messages on behalf of the op sender.
func (k Keeper) payBundlerFallback(ctx context.Context, op *accountsv1.UserOperation) ([]*anypb.Any, error) {
	// TODO: execute in isolated context with gas limit
	responses := make([]*anypb.Any, len(op.BundlerPaymentMessages))
	sender, err := k.addressCodec.StringToBytes(op.Sender)
	if err != nil {
		return nil, err
	}
	for i, msg := range op.BundlerPaymentMessages {
		resp, err := k.untypedExecute(ctx, sender, msg)
		if err != nil {
			return nil, fmt.Errorf("failed to execute bundler payment message %d: %s", i, err.Error())
		}
		responses[i] = resp
	}
	return responses, nil
}

// untypedExecute executes a protobuf message without knowing the response type.
// It will check if the sender is allowed to execute the message and then execute it.
func (k Keeper) untypedExecute(ctx context.Context, sender []byte, anyMsg *anypb.Any) (*anypb.Any, error) {
	msg, err := anyMsg.UnmarshalNew()
	if err != nil {
		return nil, err
	}
	// we check if the sender is allowed to execute the message.
	wantSender, err := k.getSenderFunc(msg)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(sender, wantSender) {
		return nil, fmt.Errorf("sender %s is not allowed to execute message %T", sender, msg)
	}
	// we now need to fetch the response type from the request message type.
	// this is because the response type is not known.
	respName := k.msgResponseFromRequestName(string(msg.ProtoReflect().Descriptor().FullName()))
	if respName == "" {
		return nil, fmt.Errorf("could not find response type for message %T", msg)
	}
	// get response type
	respType, err := protoregistry.GlobalTypes.FindMessageByName(protoreflect.FullName(respName))
	if err != nil {
		return nil, err
	}
	resp := respType.New().Interface()
	err = k.execModuleFunc(ctx, msg.(protoiface.MessageV1), resp.(protoiface.MessageV1))
	if err != nil {
		return nil, err
	}
	return anypb.New(resp)
}

// parsePayBundlerResponse parses the bundler response as any into a slice of
// responses on payment messages.
func parsePayBundlerResponse(resp any) ([]*anypb.Any, error) {
	payBundlerResp, ok := resp.(*account_abstractionv1.MsgPayBundlerResponse)
	// this means the account does not properly implement account abstraction.
	if !ok {
		return nil, fmt.Errorf("account does not implement account abstraction correctly: wanted %T, got %T", &account_abstractionv1.MsgPayBundlerResponse{}, resp)
	}
	return payBundlerResp.BundlerPaymentMessagesResponse, nil
}
