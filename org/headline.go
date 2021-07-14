package org

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

type Outline struct {
	*Section
	last  *Section
	count int
}

type Section struct {
	Headline *Headline
	Parent   *Section
	Children []*Section
}

type Headline struct {
	Index      int
	Lvl        int
	Status     string
	Priority   string
	Properties *PropertyDrawer
	Id         string
	Text       string
	Title      []Node
	Tags       []string
	Children   []Node
}

var headlineRegexp = regexp.MustCompile(`^([*]+)\s+(.*)`)
var tagRegexp = regexp.MustCompile(`(.*?)\s+(:[A-Za-z0-9_@#%:]+:\s*$)`)

func lexHeadline(line string) (token, bool) {
	if m := headlineRegexp.FindStringSubmatch(line); m != nil {
		return token{"headline", 0, m[2], m}, true
	}
	return nilToken, false
}

func (d *Document) parseHeadline(i int, parentStop stopFn) (int, Node) {
	t, headline := d.tokens[i], Headline{}
	headline.Lvl = len(t.matches[1])

	headline.Index = d.addHeadline(&headline)

	text := t.content
	todoKeywords := strings.FieldsFunc(d.Get("TODO"), func(r rune) bool { return unicode.IsSpace(r) || r == '|' })
	for _, k := range todoKeywords {
		if strings.HasPrefix(text, k) && len(text) > len(k) && unicode.IsSpace(rune(text[len(k)])) {
			headline.Status = k
			text = text[len(k)+1:]
			break
		}
	}

	if len(text) >= 4 && text[0:2] == "[#" && strings.Contains("ABC", text[2:3]) && text[3] == ']' {
		headline.Priority = text[2:3]
		text = strings.TrimSpace(text[4:])
	}

	if m := tagRegexp.FindStringSubmatch(text); m != nil {
		text = m[1]
		headline.Tags = strings.FieldsFunc(m[2], func(r rune) bool { return r == ':' })
	}

	headline.Title = d.parseInline(text)
	headline.Text = String(headline.Title)

	stop := func(d *Document, i int) bool {
		return parentStop(d, i) || d.tokens[i].kind == "headline" && len(d.tokens[i].matches[1]) <= headline.Lvl
	}
	consumed, nodes := d.parseMany(i+1, stop)
	if len(nodes) > 0 {
		if d, ok := nodes[0].(PropertyDrawer); ok {
			headline.Properties = &d
			nodes = nodes[1:]
		}
	}
	headline.Children = nodes
	return consumed + 1, headline
}

func (h Headline) ID(d *Document) string {
	if 0 != len(h.Id) {
		return  h.Id
	}
	if customID, ok := h.Properties.Get("CUSTOM_ID"); ok {
		h.Id = customID
		d.Ids[h.Id] = true
		return h.Id
	}
	id := strings.ToLower(h.Text)
	// Loosely based on https://github.com/bryanbraun/anchorjs/blob/master/anchor.js
        // I am not at all sure this regexp is sufficient.
	id = regexp.MustCompile("[][& +$,:;=?@\"#{}|^~`%!'<>./()*\\\n\t\b\v\u00A0]+").ReplaceAllString(id, "-")
	id = regexp.MustCompile(`-+`).ReplaceAllString(id, "-")
	id = strings.TrimPrefix(id, "-")
	id = strings.TrimSuffix(id, "-")
	base := id
	count := 0
	for {
		if d.Ids[id] {
			count++
			id = fmt.Sprintf("%s-%d", base, count)
		} else {
			break
		}
	}
	h.Id = id
	d.Ids[h.Id] = true
	return h.Id
}

func (h Headline) IsExcluded(d *Document) bool {
	for _, excludedTag := range strings.Fields(d.Get("EXCLUDE_TAGS")) {
		for _, tag := range h.Tags {
			if tag == excludedTag {
				return true
			}
		}
	}
	return false
}

func (parent *Section) add(current *Section) {
	if parent.Headline == nil || parent.Headline.Lvl < current.Headline.Lvl {
		parent.Children = append(parent.Children, current)
		current.Parent = parent
	} else {
		parent.Parent.add(current)
	}
}

func (n Headline) String() string { return orgWriter.WriteNodesAsString(n) }
