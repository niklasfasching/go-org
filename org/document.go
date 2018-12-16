package org

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"strings"
)

type Document struct {
	Path                string
	tokens              []token
	Nodes               []Node
	Footnotes           *Footnotes
	StatusKeywords      []string
	MaxEmphasisNewLines int
	AutoLink            bool
	BufferSettings      map[string]string
	DefaultSettings     map[string]string
	Error               error
}

type Writer interface {
	before(*Document)
	after(*Document)
	writeNodes(...Node)
	String() string
}

type Node interface{}

type lexFn = func(line string) (t token, ok bool)
type parseFn = func(*Document, int, stopFn) (int, Node)
type stopFn = func(*Document, int) bool

type token struct {
	kind    string
	lvl     int
	content string
	matches []string
}

var lexFns = []lexFn{
	lexHeadline,
	lexBlock,
	lexList,
	lexTable,
	lexHorizontalRule,
	lexKeywordOrComment,
	lexFootnoteDefinition,
	lexExample,
	lexText,
}

var nilToken = token{"nil", -1, "", nil}

var DefaultFrontMatterHandler = func(k, v string) interface{} {
	switch k {
	case "TAGS":
		return strings.Fields(v)
	default:
		return v
	}
}

func NewDocument() *Document {
	return &Document{
		Footnotes: &Footnotes{
			ExcludeHeading: true,
			Title:          "Footnotes",
			Definitions:    map[string]*FootnoteDefinition{},
		},
		AutoLink:            true,
		MaxEmphasisNewLines: 1,
		BufferSettings:      map[string]string{},
		DefaultSettings: map[string]string{
			"TODO": "TODO | DONE",
		},
	}
}

func (d *Document) Write(w Writer) (out string, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("could not write output: %s", recovered)
		}
	}()
	if d.Error != nil {
		return "", d.Error
	} else if d.Nodes == nil {
		return "", fmt.Errorf("could not write output: parse was not called")
	}
	w.before(d)
	w.writeNodes(d.Nodes...)
	w.after(d)
	return w.String(), err
}

func (d *Document) Parse(input io.Reader) *Document {
	defer func() {
		if err := recover(); err != nil {
			d.Error = fmt.Errorf("could not parse input: %s", err)
		}
	}()
	if d.tokens != nil {
		d.Error = fmt.Errorf("parse was called multiple times")
	}
	d.tokenize(input)
	_, nodes := d.parseMany(0, func(d *Document, i int) bool { return !(i < len(d.tokens)) })
	d.Nodes = nodes
	return d
}

func (d *Document) SetPath(path string) *Document {
	d.Path = path
	return d
}

func (d *Document) FrontMatter(input io.Reader, f func(string, string) interface{}) (_ map[string]interface{}, err error) {
	defer func() {
		d.tokens = nil
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("could not parse input: %s", recovered)
		}
	}()
	d.tokenize(input)
	d.parseMany(0, func(d *Document, i int) bool {
		if !(i < len(d.tokens)) {
			return true
		}
		t := d.tokens[i]
		return t.kind != "keyword" && !(t.kind == "text" && t.content == "")
	})
	frontMatter := make(map[string]interface{}, len(d.BufferSettings))
	for k, v := range d.BufferSettings {
		frontMatter[k] = f(k, v)
	}
	return frontMatter, err
}

func (d *Document) tokenize(input io.Reader) {
	d.tokens = []token{}
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		d.tokens = append(d.tokens, tokenize(scanner.Text()))
	}
	if err := scanner.Err(); err != nil {
		d.Error = fmt.Errorf("could not tokenize input: %s", err)
	}
}

func (d *Document) Get(key string) string {
	if v, ok := d.BufferSettings[key]; ok {
		return v
	}
	if v, ok := d.DefaultSettings[key]; ok {
		return v
	}
	return ""
}

func (d *Document) parseOne(i int, stop stopFn) (consumed int, node Node) {
	switch d.tokens[i].kind {
	case "unorderedList", "orderedList":
		consumed, node = d.parseList(i, stop)
	case "tableRow", "tableSeparator":
		consumed, node = d.parseTable(i, stop)
	case "beginBlock":
		consumed, node = d.parseBlock(i, stop)
	case "text":
		consumed, node = d.parseParagraph(i, stop)
	case "example":
		consumed, node = d.parseExample(i, stop)
	case "horizontalRule":
		consumed, node = d.parseHorizontalRule(i, stop)
	case "comment":
		consumed, node = d.parseComment(i, stop)
	case "keyword":
		consumed, node = d.parseKeyword(i, stop)
	case "headline":
		consumed, node = d.parseHeadline(i, stop)
	case "footnoteDefinition":
		consumed, node = d.parseFootnoteDefinition(i, stop)
	}

	if consumed != 0 {
		return consumed, node
	}
	log.Printf("Could not parse token %#v: Falling back to treating it as plain text.", d.tokens[i])
	m := plainTextRegexp.FindStringSubmatch(d.tokens[i].matches[0])
	d.tokens[i] = token{"text", len(m[1]), m[2], m}
	return d.parseOne(i, stop)
}

func (d *Document) parseMany(i int, stop stopFn) (int, []Node) {
	start, nodes := i, []Node{}
	for i < len(d.tokens) && !stop(d, i) {
		consumed, node := d.parseOne(i, stop)
		i += consumed
		nodes = append(nodes, node)
	}
	return i - start, nodes
}

func tokenize(line string) token {
	for _, lexFn := range lexFns {
		if token, ok := lexFn(line); ok {
			return token
		}
	}
	panic(fmt.Sprintf("could not lex line: %s", line))
}
