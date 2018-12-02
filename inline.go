package org

import (
	"regexp"
	"strings"
	"unicode"
)

type Text struct{ Content string }

type Linebreak struct{}

type Emphasis struct {
	Kind    string
	Content []Node
}

type FootnoteLink struct{ Name string }

type RegularLink struct {
	Protocol    string
	Description []Node
	URL         string
}

var redundantSpaces = regexp.MustCompile("[ \t]+")
var subScriptSuperScriptRegexp = regexp.MustCompile(`([_^])\{(.*?)\}`)
var footnoteRegexp = regexp.MustCompile(`\[fn:([\w-]+?)(:(.*?))?\]`)

func (d *Document) parseInline(input string) (nodes []Node) {
	previous, current := 0, 0
	for current < len(input) {
		consumed, node := 0, (Node)(nil)
		switch input[current] {
		case '^':
			consumed, node = d.parseSubOrSuperScript(input, current)
		case '_':
			consumed, node = d.parseSubScriptOrEmphasis(input, current)
		case '*', '/', '=', '~', '+':
			consumed, node = d.parseEmphasis(input, current)
		case '[':
			consumed, node = d.parseRegularLinkOrFootnoteReference(input, current)
		case '\\':
			consumed, node = d.parseExplicitLineBreak(input, current)
		}
		if consumed != 0 {
			if current > previous {
				nodes = append(nodes, Text{input[previous:current]})
			}
			if node != nil {
				nodes = append(nodes, node)
			}
			current += consumed
			previous = current
		} else {
			current++
		}
	}

	if previous < len(input) {
		nodes = append(nodes, Text{input[previous:]})
	}
	return nodes
}

func (d *Document) parseExplicitLineBreak(input string, start int) (int, Node) {
	if start == 0 || input[start-1] == '\n' || start+1 >= len(input) || input[start+1] != '\\' {
		return 0, nil
	}
	for i := start + 1; ; i++ {
		if i == len(input)-1 || input[i] == '\n' {
			return i + 1 - start, Linebreak{}
		}
		if !unicode.IsSpace(rune(input[i])) {
			break
		}
	}
	return 0, nil
}

func (d *Document) parseSubOrSuperScript(input string, start int) (int, Node) {
	if m := subScriptSuperScriptRegexp.FindStringSubmatch(input[start:]); m != nil {
		return len(m[2]) + 3, Emphasis{m[1] + "{}", []Node{Text{m[2]}}}
	}
	return 0, nil
}

func (d *Document) parseSubScriptOrEmphasis(input string, start int) (int, Node) {
	if consumed, node := d.parseSubOrSuperScript(input, start); consumed != 0 {
		return consumed, node
	}
	return d.parseEmphasis(input, start)
}

func (d *Document) parseRegularLinkOrFootnoteReference(input string, start int) (int, Node) {
	if len(input[start:]) >= 2 && input[start] == '[' && input[start+1] == '[' {
		return d.parseRegularLink(input, start)
	} else if len(input[start:]) >= 1 && input[start] == '[' {
		return d.parseFootnoteReference(input, start)
	}
	return 0, nil
}

func (d *Document) parseFootnoteReference(input string, start int) (int, Node) {
	if m := footnoteRegexp.FindStringSubmatch(input[start:]); m != nil {
		name, definition := m[1], m[3]
		seen := false
		for _, otherName := range d.Footnotes.Order {
			if name == otherName {
				seen = true
			}
		}
		if !seen {
			d.Footnotes.Order = append(d.Footnotes.Order, name)
		}
		if definition != "" {
			d.Footnotes.Definitions[name] = FootnoteDefinition{name, d.parseInline(definition)}
		}
		return len(m[0]), FootnoteLink{name}
	}
	return 0, nil
}

func (d *Document) parseRegularLink(input string, start int) (int, Node) {
	if len(input[start:]) == 0 || input[start+1] != '[' {
		return 0, nil
	}
	input = input[start:]
	end := strings.Index(input, "]]")
	if end == -1 {
		return 0, nil
	}

	rawLink := input[2:end]
	link, description, parts := "", []Node{}, strings.Split(rawLink, "][")
	if len(parts) == 2 {
		link, description = parts[0], d.parseInline(parts[1])
	} else {
		link, description = rawLink, []Node{Text{rawLink}}
	}
	consumed := end + 2
	protocol, parts := "", strings.SplitN(link, ":", 2)
	if len(parts) == 2 {
		protocol = parts[0]
	}
	return consumed, RegularLink{protocol, description, link}
}

func (d *Document) parseEmphasis(input string, start int) (int, Node) {
	marker, i := input[start], start
	if !hasValidPreAndBorderChars(input, i) {
		return 0, nil
	}
	for i, consumedNewLines := i+1, 0; i < len(input) && consumedNewLines <= d.MaxEmphasisNewLines; i++ {
		if input[i] == '\n' {
			consumedNewLines++
		}

		if input[i] == marker && i != start+1 && hasValidPostAndBorderChars(input, i) {
			return i + 1 - start, Emphasis{input[start : start+1], d.parseInline(input[start+1 : i])}
		}
	}
	return 0, nil
}

// see org-emphasis-regexp-components (emacs elisp variable)

func hasValidPreAndBorderChars(input string, i int) bool {
	return (i+1 >= len(input) || isValidBorderChar(rune(input[i+1]))) && (i == 0 || isValidPreChar(rune(input[i-1])))
}

func hasValidPostAndBorderChars(input string, i int) bool {
	return (i == 0 || isValidBorderChar(rune(input[i-1]))) && (i+1 >= len(input) || isValidPostChar(rune(input[i+1])))
}

func isValidPreChar(r rune) bool {
	return unicode.IsSpace(r) || strings.ContainsRune(`-({'"`, r)
}

func isValidPostChar(r rune) bool {
	return unicode.IsSpace(r) || strings.ContainsRune(`-.,:!?;'")}[`, r)
}

func isValidBorderChar(r rune) bool { return !unicode.IsSpace(r) }
