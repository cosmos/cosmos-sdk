package cli

import (
	"github.com/spf13/cobra"
)

type RootCommand struct{ *cobra.Command }
type QueryCommand struct{ *cobra.Command }
type TxCommand struct{ *cobra.Command }

func (RootCommand) IsAutoGroupType()  {}
func (QueryCommand) IsAutoGroupType() {}
func (TxCommand) IsAutoGroupType()    {}
