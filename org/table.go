package org

import (
	"regexp"
	"strings"
)

type Table struct {
	Header Node
	Rows   []Node
}

type TableSeparator struct{ Content string }

type TableHeader struct {
	SeparatorBefore Node
	Columns         [][]Node
	SeparatorAfter  Node
}

type TableRow struct{ Columns [][]Node }

var tableSeparatorRegexp = regexp.MustCompile(`^(\s*)(\|[+-|]*)\s*$`)
var tableRowRegexp = regexp.MustCompile(`^(\s*)(\|.*)`)

func lexTable(line string) (token, bool) {
	if m := tableSeparatorRegexp.FindStringSubmatch(line); m != nil {
		return token{"tableSeparator", len(m[1]), m[2], m}, true
	} else if m := tableRowRegexp.FindStringSubmatch(line); m != nil {
		return token{"tableRow", len(m[1]), m[2], m}, true
	}
	return nilToken, false
}

func (d *Document) parseTable(i int, parentStop stopFn) (int, Node) {
	rows, start := []Node{}, i
	for !parentStop(d, i) && (d.tokens[i].kind == "tableRow" || d.tokens[i].kind == "tableSeparator") {
		consumed, row := d.parseTableRowOrSeparator(i, parentStop)
		i += consumed
		rows = append(rows, row)
	}

	consumed := i - start

	if len(rows) >= 2 {
		if row, ok := rows[0].(TableRow); ok {
			if separator, ok := rows[1].(TableSeparator); ok {
				return consumed, Table{TableHeader{nil, row.Columns, separator}, rows[2:]}
			}
		}
	}
	if len(rows) >= 3 {
		if separatorBefore, ok := rows[0].(TableSeparator); ok {
			if row, ok := rows[1].(TableRow); ok {
				if separatorAfter, ok := rows[2].(TableSeparator); ok {
					return consumed, Table{TableHeader{separatorBefore, row.Columns, separatorAfter}, rows[3:]}
				}
			}
		}
	}

	return consumed, Table{nil, rows}
}

func (d *Document) parseTableRowOrSeparator(i int, _ stopFn) (int, Node) {
	if d.tokens[i].kind == "tableSeparator" {
		return 1, TableSeparator{d.tokens[i].content}
	}
	fields := strings.FieldsFunc(d.tokens[i].content, func(r rune) bool { return r == '|' })
	row := TableRow{}
	for _, field := range fields {
		row.Columns = append(row.Columns, d.parseInline(strings.TrimSpace(field)))
	}
	return 1, row
}
