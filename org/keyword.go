package org

import (
	"regexp"
	"strings"
)

type Keyword struct {
	Key   string
	Value string
}

type NodeWithMeta struct {
	Node Node
	Meta map[string]string
}

type Comment struct{ Content string }

var keywordRegexp = regexp.MustCompile(`^(\s*)#\+([^:]+):(\s+(.*)|(\s*)$)`)
var commentRegexp = regexp.MustCompile(`^(\s*)#(.*)`)
var affiliatedKeywordRegexp = regexp.MustCompile(`^(CAPTION)$`)

func lexKeywordOrComment(line string) (token, bool) {
	if m := keywordRegexp.FindStringSubmatch(line); m != nil {
		return token{"keyword", len(m[1]), m[2], m}, true
	} else if m := commentRegexp.FindStringSubmatch(line); m != nil {
		return token{"comment", len(m[1]), m[2], m}, true
	}
	return nilToken, false
}

func (d *Document) parseKeyword(i int, stop stopFn) (int, Node) {
	k := parseKeyword(d.tokens[i])
	if affiliatedKeywordRegexp.MatchString(k.Key) {
		consumed, node := d.parseAffiliated(i, stop)
		if consumed != 0 {
			return consumed, node
		}
	} else {
		d.BufferSettings[k.Key] = strings.Join([]string{d.BufferSettings[k.Key], k.Value}, "\n")
	}
	return 1, k
}

func (d *Document) parseComment(i int, stop stopFn) (int, Node) {
	return 1, Comment{d.tokens[i].content}
}

func (d *Document) parseAffiliated(i int, stop stopFn) (int, Node) {
	start, meta := i, map[string]string{}
	for ; !stop(d, i) && d.tokens[i].kind == "keyword"; i++ {
		k := parseKeyword(d.tokens[i])
		if !affiliatedKeywordRegexp.MatchString(k.Key) {
			return 0, nil
		}
		if value, ok := meta[k.Key]; ok {
			meta[k.Key] = value + " " + k.Value
		} else {
			meta[k.Key] = k.Value
		}
	}
	if stop(d, i) {
		return 0, nil
	}
	consumed, node := 0, (Node)(nil)
	if t := d.tokens[i]; t.kind == "text" {
		if nodes := d.parseInline(t.content); len(nodes) == 1 && isImageOrVideoLink(nodes[0]) {
			consumed, node = 1, Paragraph{nodes[:1]}
		}
	} else {
		consumed, node = d.parseOne(i, stop)
	}
	if consumed == 0 || node == nil {
		return 0, nil
	}
	i += consumed
	return i - start, NodeWithMeta{node, meta}
}

func parseKeyword(t token) Keyword {
	k, v := t.matches[2], t.matches[4]
	k = strings.ToUpper(k)
	return Keyword{k, v}
}
