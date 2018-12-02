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

func (d *Document) parseBlock(i int, parentStop stopFn) (int, Node) {
	t, start, nodes := d.tokens[i], i, []Node{}
	name, parameters := t.content, strings.Fields(t.matches[3])
	trim := trimIndentUpTo(d.tokens[i].lvl)
	for i++; !(d.tokens[i].kind == "endBlock" && d.tokens[i].content == name); i++ {
		if parentStop(d, i) {
			return 0, nil
		}
		nodes = append(nodes, Line{[]Node{Text{trim(d.tokens[i].matches[0])}}})
	}
	return i + 1 - start, Block{name, parameters, nodes}
}

func trimIndentUpTo(max int) func(string) string {
	return func(line string) string {
		i := 0
		for ; i < len(line) && i < max && unicode.IsSpace(rune(line[i])); i++ {
		}
		return line[i:]
	}
}
