package org

import (
	"regexp"
)

type Footnotes struct {
	ExcludeHeading bool
	Title          string
	Definitions    map[string]*FootnoteDefinition
	addOrder       []string
}

type FootnoteDefinition struct {
	Name     string
	Children []Node
	Inline   bool
}

var footnoteDefinitionRegexp = regexp.MustCompile(`^\[fn:([\w-]+)\](\s+(.+)|$)`)

func lexFootnoteDefinition(line string) (token, bool) {
	if m := footnoteDefinitionRegexp.FindStringSubmatch(line); m != nil {
		return token{"footnoteDefinition", 0, m[1], m}, true
	}
	return nilToken, false
}

func (d *Document) parseFootnoteDefinition(i int, parentStop stopFn) (int, Node) {
	start, name := i, d.tokens[i].content
	d.tokens[i] = tokenize(d.tokens[i].matches[2])
	stop := func(d *Document, i int) bool {
		return parentStop(d, i) ||
			(isSecondBlankLine(d, i) && i > start+1) ||
			d.tokens[i].kind == "headline" || d.tokens[i].kind == "footnoteDefinition"
	}
	consumed, nodes := d.parseMany(i, stop)
	d.Footnotes.add(name, &FootnoteDefinition{name, nodes, false})
	return consumed, nil
}

func (fs *Footnotes) add(name string, definition *FootnoteDefinition) {
	if definition != nil {
		fs.Definitions[name] = definition
	}
	fs.addOrder = append(fs.addOrder, name)
}

func (fs *Footnotes) Ordered() []FootnoteDefinition {
	m := map[string]bool{}
	definitions, inlineDefinitions := []FootnoteDefinition{}, []FootnoteDefinition{}
	for _, name := range fs.addOrder {
		if isDuplicate := m[name]; !isDuplicate {
			m[name] = true
			if definition := *fs.Definitions[name]; definition.Inline {
				inlineDefinitions = append(inlineDefinitions, definition)
			} else {
				definitions = append(definitions, definition)
			}
		}
	}
	return append(definitions, inlineDefinitions...)
}
