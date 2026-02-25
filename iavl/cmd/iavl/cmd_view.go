package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

func newViewCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "view [dir]",
		Aliases: []string{"v"},
		Short:   "Interactively browse IAVL store data",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) > 0 {
				dir = args[0]
			}
			p := tea.NewProgram(initialModel(dir), tea.WithAltScreen())
			_, err := p.Run()
			return err
		},
	}
}
