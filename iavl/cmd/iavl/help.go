package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	"github.com/cosmos/cosmos-sdk/iavl/docs"
)

// helpDocer is optionally implemented by views to provide per-view documentation.
type helpDocer interface {
	HelpDoc() string
}

// helpModal holds state for the floating help documentation modal.
type helpModal struct {
	visible  bool
	viewport viewport.Model
	content  string // raw markdown, stored for re-render on resize
	width    int
	height   int
}

func (h *helpModal) open(doc string, totalW, totalH int) {
	h.content = doc
	h.width = totalW
	h.height = totalH
	h.visible = true

	// Border is 1 cell on each side; padding adds 2 more each side horizontally.
	// vpW = totalW - 2 (border) - 4 (padding)
	vpW := totalW - 6
	if vpW < 10 {
		vpW = 10
	}
	// border (2) + title (1) + footer (1) = 4 lines overhead
	vpH := totalH - 4
	if vpH < 5 {
		vpH = 5
	}

	r, err := glamour.NewTermRenderer(glamour.WithAutoStyle(), glamour.WithWordWrap(vpW))
	var rendered string
	if err == nil {
		rendered, err = r.Render(doc)
	}
	if err != nil {
		rendered = doc
	}

	vp := viewport.New(vpW, vpH)
	vp.SetContent(rendered)
	h.viewport = vp
}

func (h *helpModal) resize(totalW, totalH int) {
	if !h.visible {
		return
	}
	h.open(h.content, totalW, totalH)
}

func (h *helpModal) render(totalW, totalH int) string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("62"))
	footerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("242"))

	title := titleStyle.Render(" Documentation ")
	body := h.viewport.View()
	footer := footerStyle.Render(" ? or esc  close  •  ↑/↓ / j/k  scroll  •  pgup/pgdn  page ")

	content := title + "\n" + body + "\n" + footer

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(0, 2).
		Width(totalW - 2)

	return boxStyle.Render(content)
}

// Per-view help doc constants.

const appHelpDoc = `# IAVL Browser

Interactive terminal UI for browsing IAVL store data.

## Navigation

- **enter** — drill into selected item
- **esc** — go back to previous view
- **q** — quit
- **r** — refresh current view
- **?** — open this documentation

## Views

| View | Description |
|------|-------------|
| Trees | Top-level list of all IAVL trees in the store directory |
| Changesets | Changesets (version segments) for a single tree |
| Checkpoints | Checkpoints within a changeset |
| Leaves | Leaf nodes in a checkpoint |
| Branches | Branch nodes in a checkpoint |
| Orphans | Orphaned nodes in a changeset |
| WAL Analysis | Write-ahead log summary by version |
| WAL Entries | Individual WAL entries for one version |
| Commit Info | Committed block hashes and store metadata |

## Table Controls

- **↑ / k** — move up
- **↓ / j** — move down
- **pgup / pgdn** — page up / down
- **home / end** — jump to top / bottom
`

func loadDocMarkdown(docFileName string) string {
	data, err := docs.Docs.ReadFile(docFileName)
	if err != nil {
		return fmt.Sprintf("# %s\n\nDocumentation not found: %s", docFileName, err)
	}
	return string(data)
}

const treesHelpDoc = "# Trees\n\nTODO: trees docs\n"

const checkpointsHelpDoc = "# Checkpoints\n\nTODO: checkpoints docs\n"

const leavesHelpDoc = "# Leaves\n\nTODO: leaves docs\n"

const branchesHelpDoc = "# Branches\n\nTODO: branches docs\n"

const orphansHelpDoc = "# Orphans\n\nTODO: orphans docs\n"

const walAnalysisHelpDoc = "# WAL Analysis\n\nTODO: WAL analysis docs\n"

const walEntriesHelpDoc = "# WAL Entries\n\nTODO: WAL entries docs\n"

const commitInfoHelpDoc = "# Commit Info\n\nTODO: commit info docs\n"
