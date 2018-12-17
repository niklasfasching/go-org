package org

import (
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
)

type Comment struct{ Content string }

type Keyword struct {
	Key   string
	Value string
}

type NodeWithMeta struct {
	Node Node
	Meta Metadata
}

type Metadata struct {
	Caption        [][]Node
	HTMLAttributes [][]string
}

type Include struct {
	Keyword
	Resolve func() Node
}

var keywordRegexp = regexp.MustCompile(`^(\s*)#\+([^:]+):(\s+(.*)|(\s*)$)`)
var commentRegexp = regexp.MustCompile(`^(\s*)#(.*)`)

var includeFileRegexp = regexp.MustCompile(`(?i)^"([^"]+)" (src|example|export) (\w+)$`)

func lexKeywordOrComment(line string) (token, bool) {
	if m := keywordRegexp.FindStringSubmatch(line); m != nil {
		return token{"keyword", len(m[1]), m[2], m}, true
	} else if m := commentRegexp.FindStringSubmatch(line); m != nil {
		return token{"comment", len(m[1]), m[2], m}, true
	}
	return nilToken, false
}

func (d *Document) parseComment(i int, stop stopFn) (int, Node) {
	return 1, Comment{d.tokens[i].content}
}

func (d *Document) parseKeyword(i int, stop stopFn) (int, Node) {
	k := parseKeyword(d.tokens[i])
	switch k.Key {
	case "INCLUDE":
		return d.newInclude(k)
	case "CAPTION", "ATTR_HTML":
		consumed, node := d.parseAffiliated(i, stop)
		if consumed != 0 {
			return consumed, node
		}
		fallthrough
	default:
		if _, ok := d.BufferSettings[k.Key]; ok {
			d.BufferSettings[k.Key] = strings.Join([]string{d.BufferSettings[k.Key], k.Value}, "\n")
		} else {
			d.BufferSettings[k.Key] = k.Value
		}
		return 1, k
	}
}

func (d *Document) parseAffiliated(i int, stop stopFn) (int, Node) {
	start, meta := i, Metadata{}
	for ; !stop(d, i) && d.tokens[i].kind == "keyword"; i++ {
		switch k := parseKeyword(d.tokens[i]); k.Key {
		case "CAPTION":
			meta.Caption = append(meta.Caption, d.parseInline(k.Value))
		case "ATTR_HTML":
			r := csv.NewReader(strings.NewReader(k.Value))
			r.Comma = ' '
			attributes, err := r.Read()
			if err != nil {
				return 0, nil
			}
			meta.HTMLAttributes = append(meta.HTMLAttributes, attributes)
		default:
			return 0, nil
		}
	}
	if stop(d, i) {
		return 0, nil
	}
	consumed, node := d.parseOne(i, stop)
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

func (d *Document) newInclude(k Keyword) (int, Node) {
	resolve := func() Node { panic(fmt.Sprintf("bad include: '#+INCLUDE: %s'", k.Value)) }
	if m := includeFileRegexp.FindStringSubmatch(k.Value); m != nil {
		path, kind, lang := m[1], m[2], m[3]
		if !filepath.IsAbs(path) {
			path = filepath.Join(filepath.Dir(d.Path), path)
		}
		resolve = func() Node {
			bs, err := ioutil.ReadFile(path)
			if err != nil {
				panic(fmt.Sprintf("bad include '#+INCLUDE: %s': %s", k.Value, err))
			}
			return Block{strings.ToUpper(kind), []string{lang}, d.parseRawInline(string(bs))}
		}
	}
	return 1, Include{k, resolve}
}
