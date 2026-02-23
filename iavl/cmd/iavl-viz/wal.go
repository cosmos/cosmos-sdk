package main

import (
	"encoding/hex"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
)

func (m *model) buildWALAnalysisTable(info []walVersionInfo, total walVersionInfo) {
	rows := make([]table.Row, 0, len(info)+1)
	for i := range info {
		v := &info[i]
		rows = append(rows, table.Row{
			strconv.FormatUint(v.version, 10),
			strconv.Itoa(v.sets),
			strconv.Itoa(v.deletes),
			formatStat(&v.keyStats, (*runningStats).avg),
			formatStat(&v.keyStats, (*runningStats).stddev),
			formatStat(&v.valStats, (*runningStats).avg),
			formatStat(&v.valStats, (*runningStats).stddev),
			strconv.Itoa(v.offset),
		})
	}
	rows = append(rows, table.Row{
		"━━ TOTAL",
		strconv.Itoa(total.sets),
		strconv.Itoa(total.deletes),
		formatStat(&total.keyStats, (*runningStats).avg),
		formatStat(&total.keyStats, (*runningStats).stddev),
		formatStat(&total.valStats, (*runningStats).avg),
		formatStat(&total.valStats, (*runningStats).stddev),
		"━━",
	})
	cols := []table.Column{
		{Title: "Version", Width: 10},
		{Title: "Sets", Width: 8},
		{Title: "Deletes", Width: 8},
		{Title: "Avg Key", Width: 10},
		{Title: "Key StdDev", Width: 10},
		{Title: "Avg Val", Width: 10},
		{Title: "Val StdDev", Width: 10},
		{Title: "Offset", Width: 10},
	}
	m.columns = cols
	m.table = newTable(cols, rows, m.tableHeight())
}

func (m *model) buildWALEntriesTable(entries []walEntry) {
	rows := make([]table.Row, len(entries))
	for i := range entries {
		e := &entries[i]
		del := "no"
		if e.delete {
			del = "yes"
		}
		rows[i] = table.Row{
			hex.EncodeToString(e.key),
			hex.EncodeToString(e.value),
			del,
		}
	}
	cols := []table.Column{
		{Title: "Key", Width: 50},
		{Title: "Value", Width: 50},
		{Title: "Delete", Width: 8},
	}
	m.columns = cols
	m.table = newTable(cols, rows, m.tableHeight())
}
