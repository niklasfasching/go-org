package org

import (
	"regexp"
	"strings"
)

type Keyword struct {
	Key   string
	Value string
}

type Comment struct{ Content string }

var keywordRegexp = regexp.MustCompile(`^(\s*)#\+([^:]+):\s(.*)`)
var commentRegexp = regexp.MustCompile(`^(\s*)#(.*)`)

func lexKeywordOrComment(line string) (token, bool) {
	if m := keywordRegexp.FindStringSubmatch(line); m != nil {
		return token{"keyword", len(m[1]), m[2], m}, true
	} else if m := commentRegexp.FindStringSubmatch(line); m != nil {
		return token{"comment", len(m[1]), m[2], m}, true
	}
	return nilToken, false
}

func (d *Document) parseKeyword(i int, stop stopFn) (int, Node) {
	t := d.tokens[i]
	k, v := t.matches[2], t.matches[3]
	d.BufferSettings[k] = strings.Join([]string{d.BufferSettings[k], v}, "\n")
	return 1, Keyword{k, v}
}

func (d *Document) parseComment(i int, stop stopFn) (int, Node) {
	return 1, Comment{d.tokens[i].content}
}
