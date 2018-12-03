package org

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

type List struct {
	Kind  string
	Items []Node
}

type ListItem struct {
	Bullet   string
	Children []Node
}

var unorderedListRegexp = regexp.MustCompile(`^(\s*)([-]|[+]|[*])\s(.*)`)
var orderedListRegexp = regexp.MustCompile(`^(\s*)(([0-9]+|[a-zA-Z])[.)])\s+(.*)`)

func lexList(line string) (token, bool) {
	if m := unorderedListRegexp.FindStringSubmatch(line); m != nil {
		return token{"unorderedList", len(m[1]), m[3], m}, true
	} else if m := orderedListRegexp.FindStringSubmatch(line); m != nil {
		return token{"orderedList", len(m[1]), m[4], m}, true
	}
	return nilToken, false
}

func isListToken(t token) bool {
	return t.kind == "unorderedList" || t.kind == "orderedList"
}

func stopIndentBelow(t token, minIndent int) bool {
	return t.lvl < minIndent && !(t.kind == "text" && t.content == "")
}

func listKind(t token) string {
	switch bullet := t.matches[2]; {
	case bullet == "*" || bullet == "+" || bullet == "-":
		return bullet
	case unicode.IsLetter(rune(bullet[0])):
		return "letter"
	case unicode.IsDigit(rune(bullet[0])):
		return "number"
	default:
		panic(fmt.Sprintf("bad list bullet '%s': %#v", bullet, t))
	}
}

func (d *Document) parseList(i int, parentStop stopFn) (int, Node) {
	start, lvl := i, d.tokens[i].lvl

	list := List{Kind: listKind(d.tokens[i])}
	for !parentStop(d, i) && d.tokens[i].lvl == lvl && isListToken(d.tokens[i]) && listKind(d.tokens[i]) == list.Kind {
		consumed, node := d.parseListItem(i, parentStop)
		i += consumed
		list.Items = append(list.Items, node)
	}
	return i - start, list
}

func (d *Document) parseListItem(i int, parentStop stopFn) (int, Node) {
	start, nodes, bullet := i, []Node{}, d.tokens[i].matches[2]
	minIndent := d.tokens[i].lvl + len(bullet)
	d.tokens[i] = tokenize(strings.Repeat(" ", minIndent) + d.tokens[i].content)
	stop := func(d *Document, i int) bool {
		if parentStop(d, i) {
			return true
		}
		t := d.tokens[i]
		return t.lvl < minIndent && !(t.kind == "text" && t.content == "")
	}
	for !stop(d, i) && !isSecondBlankLine(d, i) {
		consumed, node := d.parseOne(i, stop)
		i += consumed
		nodes = append(nodes, node)
	}
	return i - start, ListItem{bullet, nodes}
}
