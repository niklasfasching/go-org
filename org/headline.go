package org

import (
	"regexp"
	"strings"
	"unicode"
)

type Headline struct {
	Lvl        int
	Status     string
	Priority   string
	Properties Node
	Title      []Node
	Tags       []string
	Children   []Node
}

var headlineRegexp = regexp.MustCompile(`^([*]+)\s+(.*)`)
var tagRegexp = regexp.MustCompile(`(.*?)\s*(:[A-Za-z0-9@#%:]+:\s*$)`)

func lexHeadline(line string) (token, bool) {
	if m := headlineRegexp.FindStringSubmatch(line); m != nil {
		return token{"headline", 0, m[2], m}, true
	}
	return nilToken, false
}

func (d *Document) todoKeywords() []string {
	return strings.FieldsFunc(d.Get("TODO"), func(r rune) bool {
		return unicode.IsSpace(r) || r == '|'
	})
}

func (d *Document) parseHeadline(i int, parentStop stopFn) (int, Node) {
	t, headline := d.tokens[i], Headline{}
	headline.Lvl = len(t.matches[1])
	text := t.content

	for _, k := range d.todoKeywords() {
		if strings.HasPrefix(text, k) && len(text) > len(k) && unicode.IsSpace(rune(text[len(k)])) {
			headline.Status = k
			text = text[len(k)+1:]
			break
		}
	}

	if len(text) >= 3 && text[0:2] == "[#" && strings.Contains("ABC", text[2:3]) && text[3] == ']' {
		headline.Priority = text[2:3]
		text = strings.TrimSpace(text[4:])
	}

	if m := tagRegexp.FindStringSubmatch(text); m != nil {
		text = m[1]
		headline.Tags = strings.FieldsFunc(m[2], func(r rune) bool { return r == ':' })
	}

	headline.Title = d.parseInline(text)

	stop := func(d *Document, i int) bool {
		return parentStop(d, i) || d.tokens[i].kind == "headline" && d.tokens[i].lvl <= headline.Lvl
	}
	consumed, nodes := d.parseMany(i+1, stop)
	if len(nodes) > 0 {
		if d, ok := nodes[0].(Drawer); ok && d.Name == "PROPERTIES" {
			headline.Properties = d
			nodes = nodes[1:]
		}
	}
	headline.Children = nodes
	return consumed + 1, headline
}
