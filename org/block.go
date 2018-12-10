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

func isRawTextBlock(name string) bool { return name == "SRC" || name == "EXAMPLE" }

func (d *Document) parseBlock(i int, parentStop stopFn) (int, Node) {
	t, start, lines := d.tokens[i], i, []string{}
	name, parameters := t.content, strings.Fields(t.matches[3])
	trim := trimIndentUpTo(d.tokens[i].lvl)
	stop := func(d *Document, i int) bool {
		return parentStop(d, i) || (d.tokens[i].kind == "endBlock" && d.tokens[i].content == name)
	}
	block, consumed, i := Block{name, parameters, nil}, 0, i+1
	if isRawTextBlock(name) {
		for ; !stop(d, i); i++ {
			lines = append(lines, trim(d.tokens[i].matches[0]))
		}
		consumed = i - start
		block.Children = []Node{Text{strings.Join(lines, "\n")}}
	} else {
		consumed, block.Children = d.parseMany(i, stop)
		consumed++ // line with BEGIN
	}
	if parentStop(d, i) {
		return 0, nil
	}
	return consumed + 1, block
}

func trimIndentUpTo(max int) func(string) string {
	return func(line string) string {
		i := 0
		for ; i < len(line) && i < max && unicode.IsSpace(rune(line[i])); i++ {
		}
		return line[i:]
	}
}
