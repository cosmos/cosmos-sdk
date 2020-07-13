package client

import (
	"bytes"
	"context"
	"fmt"

	"github.com/spf13/cobra"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TestAccountRetriever is an AccountRetriever that can be used in unit tests
type TestAccountRetriever struct {
	Accounts map[string]struct {
		Address sdk.AccAddress
		Num     uint64
		Seq     uint64
	}
}

var _ AccountRetriever = TestAccountRetriever{}

// EnsureExists implements AccountRetriever.EnsureExists
func (t TestAccountRetriever) EnsureExists(_ NodeQuerier, addr sdk.AccAddress) error {
	_, ok := t.Accounts[addr.String()]
	if !ok {
		return fmt.Errorf("account %s not found", addr)
	}
	return nil
}

// GetAccountNumberSequence implements AccountRetriever.GetAccountNumberSequence
func (t TestAccountRetriever) GetAccountNumberSequence(_ NodeQuerier, addr sdk.AccAddress) (accNum uint64, accSeq uint64, err error) {
	acc, ok := t.Accounts[addr.String()]
	if !ok {
		return 0, 0, fmt.Errorf("account %s not found", addr)
	}
	return acc.Num, acc.Seq, nil
}

// CallCliCmd calls theCmd cobra command and returns the output in bytes.
func CallCliCmd(clientCtx Context, theCmd func() *cobra.Command, extraArgs []string) ([]byte, error) {
	buf := new(bytes.Buffer)
	clientCtx = clientCtx.WithOutput(buf)

	ctx := context.Background()
	ctx = context.WithValue(ctx, ClientContextKey, &clientCtx)

	cmd := theCmd()
	cmd.SetErr(buf)
	cmd.SetOut(buf)

	cmd.SetArgs(extraArgs)

	if err := cmd.ExecuteContext(ctx); err != nil {
		return buf.Bytes(), err
	}

	return buf.Bytes(), nil
}
