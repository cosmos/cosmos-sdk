package account_abstraction

import (
	"context"
	"fmt"

	"cosmossdk.io/x/accounts/accountstd"
	account_abstractionv1 "cosmossdk.io/x/accounts/interfaces/account_abstraction/v1"
)

// FullAbstractedAccount is an account abstraction that implements
// the account abstraction interface fully. It is used for testing.
type FullAbstractedAccount struct {
	m *MinimalAbstractedAccount
}

func NewFullAbstractedAccount(d accountstd.Dependencies) (FullAbstractedAccount, error) {
	m, err := NewMinimalAbstractedAccount(d)
	if err != nil {
		return FullAbstractedAccount{}, err
	}
	return FullAbstractedAccount{m: &m}, nil
}

func (a FullAbstractedAccount) ExecuteMessages(ctx context.Context, msg *account_abstractionv1.MsgExecute) (*account_abstractionv1.MsgExecuteResponse, error) {
	// we always want to ensure that this is called by the x/accounts module, it's the only trusted entrypoint.
	// if we do not do this check then someone could call this method directly and bypass the authentication.
	if !accountstd.SenderIsAccountsModule(ctx) {
		return nil, fmt.Errorf("sender is not the x/accounts module")
	}
	// we simulate this account does not allow delegation messages to be executed.
	for _, m := range msg.ExecutionMessages {
		if m.TypeUrl == "/cosmos.staking.v1beta1.MsgDelegate" { // NOTE: this is not a safe way to check the typeUrl, it's just for testing.
			return nil, fmt.Errorf("this account does not allow delegation messages")
		}
	}
	// execute messages
	responses, err := accountstd.ExecModuleAnys(ctx, msg.ExecutionMessages)
	if err != nil {
		return nil, err
	}
	return &account_abstractionv1.MsgExecuteResponse{ExecutionMessagesResponse: responses}, nil
}

func (a FullAbstractedAccount) PayBundler(ctx context.Context, msg *account_abstractionv1.MsgPayBundler) (*account_abstractionv1.MsgPayBundlerResponse, error) {
	// we always want to ensure that this is called by the x/accounts module, it's the only trusted entrypoint.
	// if we do not do this check then someone could call this method directly and bypass the authentication.
	if !accountstd.SenderIsAccountsModule(ctx) {
		return nil, fmt.Errorf("sender is not the x/accounts module")
	}
	// we check if it's a bank send, if it is we reject it.
	for _, m := range msg.BundlerPaymentMessages {
		if m.TypeUrl == "/cosmos.bank.v1beta1.MsgSend" { // NOTE: this is not a safe way to check the typeUrl, it's just for testing.
			return nil, fmt.Errorf("this account does not allow bank send messages")
		}
	}
	// execute messages
	responses, err := accountstd.ExecModuleAnys(ctx, msg.BundlerPaymentMessages)
	if err != nil {
		return nil, err
	}
	return &account_abstractionv1.MsgPayBundlerResponse{BundlerPaymentMessagesResponse: responses}, nil
}

func (a FullAbstractedAccount) RegisterInitHandler(builder *accountstd.InitBuilder) {
	a.m.RegisterInitHandler(builder) // registers same init message as MinimalAbstractedAccount
}

func (a FullAbstractedAccount) RegisterExecuteHandlers(builder *accountstd.ExecuteBuilder) {
	accountstd.RegisterExecuteHandler(builder, a.ExecuteMessages) // implements accounts_abstraction
	accountstd.RegisterExecuteHandler(builder, a.PayBundler)      // implements account_abstraction
	a.m.RegisterExecuteHandlers(builder)                          // note: MinimalAbstractedAccount implements account_abstraction, and we're calling its RegisterExecuteHandlers
}

func (a FullAbstractedAccount) RegisterQueryHandlers(builder *accountstd.QueryBuilder) {
	a.m.RegisterQueryHandlers(builder)
}
