package main

import (
	"encoding/hex"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
)

func (m *model) buildCommitInfoTable() {
	rows := make([]table.Row, len(m.commitInfos))
	for i, ci := range m.commitInfos {
		if ci.err != nil {
			rows[i] = table.Row{ci.version, "-", "-", ci.err.Error()}
		} else {
			var storeNames []string
			for _, si := range ci.info.StoreInfos {
				storeNames = append(storeNames, si.Name)
			}
			rows[i] = table.Row{
				strconv.FormatInt(ci.info.Version, 10),
				hex.EncodeToString(ci.info.Hash()),
				strings.Join(storeNames, ", "),
				"-",
			}
		}
	}
	cols := []table.Column{
		{Title: "Version", Width: 10},
		{Title: "Hash", Width: 66},
		{Title: "Stores", Width: 50},
		{Title: "Error", Width: 30},
	}
	m.columns = cols
	m.table = newTable(cols, rows, m.tableHeight())
}
