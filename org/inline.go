package org

import (
	"path"
	"regexp"
	"strings"
	"unicode"
)

type Text struct{ Content string }

type LineBreak struct{ Count int }
type ExplicitLineBreak struct{}

type Emphasis struct {
	Kind    string
	Content []Node
}

type FootnoteLink struct {
	Name       string
	Definition *FootnoteDefinition
}

type RegularLink struct {
	Protocol    string
	Description []Node
	URL         string
	AutoLink    bool
}

var validURLCharacters = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~:/?#[]@!$&'()*+,;="
var autolinkProtocols = regexp.MustCompile(`(https?|ftp|file)`)
var imageExtensionRegexp = regexp.MustCompile(`^[.](png|gif|jpe?g|svg|tiff?)$`)
var videoExtensionRegexp = regexp.MustCompile(`^[.](webm|mp4)$`)

var subScriptSuperScriptRegexp = regexp.MustCompile(`([_^])\{(.*?)\}`)
var footnoteRegexp = regexp.MustCompile(`\[fn:([\w-]+?)(:(.*?))?\]`)

func (d *Document) parseInline(input string) (nodes []Node) {
	previous, current := 0, 0
	for current < len(input) {
		rewind, consumed, node := 0, 0, (Node)(nil)
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
		case '\n':
			consumed, node = d.parseLineBreak(input, current)
		case ':':
			rewind, consumed, node = d.parseAutoLink(input, current)
			current -= rewind
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

func (d *Document) parseLineBreak(input string, start int) (int, Node) {
	i := start
	for ; i < len(input) && input[i] == '\n'; i++ {
	}
	return i - start, LineBreak{i - start}
}

func (d *Document) parseExplicitLineBreak(input string, start int) (int, Node) {
	if start == 0 || input[start-1] == '\n' || start+1 >= len(input) || input[start+1] != '\\' {
		return 0, nil
	}
	for i := start + 2; ; i++ {
		if i == len(input)-1 || input[i] == '\n' {
			return i + 1 - start, ExplicitLineBreak{}
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
		link := FootnoteLink{name, nil}
		if definition != "" {
			link.Definition = &FootnoteDefinition{name, []Node{Paragraph{d.parseInline(definition)}}, true}
			d.Footnotes.add(name, link.Definition)
		}
		return len(m[0]), link
	}
	return 0, nil
}

func (d *Document) parseAutoLink(input string, start int) (int, int, Node) {
	if !d.AutoLink || len(input[start:]) < 3 || input[start+1] != '/' || input[start+2] != '/' {
		return 0, 0, nil
	}
	protocolStart, protocol := start-1, ""
	for ; protocolStart > 0 && unicode.IsLetter(rune(input[protocolStart])); protocolStart-- {
	}
	if m := autolinkProtocols.FindStringSubmatch(input[protocolStart:start]); m != nil {
		protocol = m[1]
	} else {
		return 0, 0, nil
	}
	end := start
	for ; end < len(input) && strings.ContainsRune(validURLCharacters, rune(input[end])); end++ {
	}
	path := input[start:end]
	if path == "://" {
		return 0, 0, nil
	}
	return len(protocol), len(path + protocol), RegularLink{protocol, nil, protocol + path, true}
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
	rawLinkParts := strings.Split(input[2:end], "][")
	description, link := ([]Node)(nil), rawLinkParts[0]
	if len(rawLinkParts) == 2 {
		link, description = rawLinkParts[0], d.parseInline(rawLinkParts[1])
	}
	consumed := end + 2
	protocol, linkParts := "", strings.SplitN(link, ":", 2)
	if len(linkParts) == 2 {
		protocol = linkParts[0]
	}
	return consumed, RegularLink{protocol, description, link, false}
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

func (l RegularLink) Kind() string {
	if p := l.Protocol; l.Description != nil || (p != "" && p != "file" && p != "http" && p != "https") {
		return "regular"
	}
	if imageExtensionRegexp.MatchString(path.Ext(l.URL)) {
		return "image"
	}
	if videoExtensionRegexp.MatchString(path.Ext(l.URL)) {
		return "video"
	}
	return "regular"
}
