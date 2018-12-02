package org

import (
	"regexp"
)

type Line struct{ Children []Node }
type Paragraph struct{ Children []Node }
type HorizontalRule struct{}

var horizontalRuleRegexp = regexp.MustCompile(`^(\s*)-{5,}\s*$`)
var plainTextRegexp = regexp.MustCompile(`^(\s*)(.*)`)

func lexText(line string) (token, bool) {
	if m := plainTextRegexp.FindStringSubmatch(line); m != nil {
		return token{"text", len(m[1]), m[2], m}, true
	}
	return nilToken, false
}

func lexHorizontalRule(line string) (token, bool) {
	if m := horizontalRuleRegexp.FindStringSubmatch(line); m != nil {
		return token{"horizontalRule", len(m[1]), "", m}, true
	}
	return nilToken, false
}

func isSecondBlankLine(d *Document, i int) bool {
	if i-1 <= 0 {
		return false
	}
	t1, t2 := d.tokens[i-1], d.tokens[i]
	if t1.kind == "text" && t2.kind == "text" && t1.content == "" && t2.content == "" {
		return true
	}
	return false
}

func (d *Document) parseParagraph(i int, parentStop stopFn) (int, Node) {
	lines, start := []Node{Line{d.parseInline(d.tokens[i].content)}}, i
	i++
	stop := func(d *Document, i int) bool { return parentStop(d, i) || d.tokens[i].kind != "text" }
	for ; !stop(d, i) && !isSecondBlankLine(d, i); i++ {
		if isSecondBlankLine(d, i) {
			lines = lines[:len(lines)-1]
			i++
			break
		}
		lines = append(lines, Line{d.parseInline(d.tokens[i].content)})
	}
	consumed := i - start
	return consumed, Paragraph{lines}
}

func (d *Document) parseHorizontalRule(i int, parentStop stopFn) (int, Node) {
	return 1, HorizontalRule{}
}
