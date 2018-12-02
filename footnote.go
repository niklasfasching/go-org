package org

import (
	"regexp"
)

type Footnotes struct {
	ExcludeHeading bool
	Title          string
	Definitions    map[string]FootnoteDefinition
	Order          []string
}

type FootnoteDefinition struct {
	Name     string
	Children []Node
}

var footnoteDefinitionRegexp = regexp.MustCompile(`^\[fn:([\w-]+)\]\s+(.+)`)

func lexFootnoteDefinition(line string) (token, bool) {
	if m := footnoteDefinitionRegexp.FindStringSubmatch(line); m != nil {
		return token{"footnoteDefinition", 0, m[1], m}, true
	}
	return nilToken, false
}

func (d *Document) parseFootnoteDefinition(i int, parentStop stopFn) (int, Node) {
	name := d.tokens[i].content
	d.tokens[i] = tokenize(d.tokens[i].matches[2])
	stop := func(d *Document, i int) bool {
		return parentStop(d, i) || isSecondBlankLine(d, i) ||
			d.tokens[i].kind == "headline" || d.tokens[i].kind == "footnoteDefinition"
	}
	consumed, nodes := d.parseMany(i, stop)
	d.Footnotes.Definitions[name] = FootnoteDefinition{name, nodes}
	return consumed, nil
}
