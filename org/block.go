package org

import (
	"regexp"
	"strings"
	"unicode"
)

type Block struct {
	Name       string
	Parameters []string
	Children   []Node
}

var beginBlockRegexp = regexp.MustCompile(`(?i)^(\s*)#\+BEGIN_(\w+)(.*)`)
var endBlockRegexp = regexp.MustCompile(`(?i)^(\s*)#\+END_(\w+)`)

func lexBlock(line string) (token, bool) {
	if m := beginBlockRegexp.FindStringSubmatch(line); m != nil {
		return token{"beginBlock", len(m[1]), strings.ToUpper(m[2]), m}, true
	} else if m := endBlockRegexp.FindStringSubmatch(line); m != nil {
		return token{"endBlock", len(m[1]), strings.ToUpper(m[2]), m}, true
	}
	return nilToken, false
}

func isRawTextBlock(name string) bool { return name == "SRC" || name == "EXAMPLE" || name == "EXPORT" }

func (d *Document) parseBlock(i int, parentStop stopFn) (int, Node) {
	t, start := d.tokens[i], i
	name, parameters := t.content, strings.Fields(t.matches[3])
	trim := trimIndentUpTo(d.tokens[i].lvl)
	stop := func(d *Document, i int) bool {
		return i >= len(d.tokens) || (d.tokens[i].kind == "endBlock" && d.tokens[i].content == name)
	}
	block, i := Block{name, parameters, nil}, i+1
	if isRawTextBlock(name) {
		rawText := ""
		for ; !stop(d, i); i++ {
			rawText += trim(d.tokens[i].matches[0]) + "\n"
		}
		block.Children = d.parseRawInline(rawText)
	} else {
		consumed, nodes := d.parseMany(i, stop)
		block.Children = nodes
		i += consumed
	}
	if i < len(d.tokens) && d.tokens[i].kind == "endBlock" && d.tokens[i].content == name {
		return i + 1 - start, block
	}
	return 0, nil
}

func trimIndentUpTo(max int) func(string) string {
	return func(line string) string {
		i := 0
		for ; i < len(line) && i < max && unicode.IsSpace(rune(line[i])); i++ {
		}
		return line[i:]
	}
}
