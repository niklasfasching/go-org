package org

import (
	"bufio"
	"fmt"
	"io"
	"log"
)

type Document struct {
	tokens              []token
	Nodes               []Node
	Footnotes           *Footnotes
	StatusKeywords      []string
	MaxEmphasisNewLines int
	BufferSettings      map[string]string
	DefaultSettings     map[string]string
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
	lexText,
}

var nilToken = token{"nil", -1, "", nil}

func NewDocument() *Document {
	return &Document{
		Footnotes: &Footnotes{
			ExcludeHeading: true,
			Title:          "Footnotes",
			Definitions:    map[string]*FootnoteDefinition{},
		},
		MaxEmphasisNewLines: 1,
		BufferSettings:      map[string]string{},
		DefaultSettings: map[string]string{
			"TODO": "TODO | DONE",
		},
	}
}

func (d *Document) Write(w Writer) Writer {
	if d.Nodes == nil {
		panic("cannot Write() empty document: you must call Parse() first")
	}
	w.before(d)
	w.writeNodes(d.Nodes...)
	w.after(d)
	return w
}

func (d *Document) Parse(input io.Reader) *Document {
	d.tokens = []token{}
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		d.tokens = append(d.tokens, tokenize(scanner.Text()))
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	_, nodes := d.parseMany(0, func(d *Document, i int) bool { return !(i < len(d.tokens)) })
	d.Nodes = nodes
	return d
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
