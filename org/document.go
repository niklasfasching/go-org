package org

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type Document struct {
	Path                string
	tokens              []token
	Nodes               []Node
	Footnotes           Footnotes
	Outline             Outline
	MaxEmphasisNewLines int
	AutoLink            bool
	BufferSettings      map[string]string
	DefaultSettings     map[string]string
	Error               error
	Log                 *log.Logger
}

type Writer interface {
	Before(*Document)
	After(*Document)
	WriteNodes(...Node)
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
	lexDrawer,
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

func NewDocument() *Document {
	outlineSection := &Section{}
	return &Document{
		Footnotes: Footnotes{
			Title:       "Footnotes",
			Definitions: map[string]*FootnoteDefinition{},
		},
		AutoLink:            true,
		MaxEmphasisNewLines: 1,
		Outline:             Outline{outlineSection, outlineSection, 0},
		BufferSettings:      map[string]string{},
		DefaultSettings: map[string]string{
			"TODO":         "TODO | DONE",
			"EXCLUDE_TAGS": "noexport",
			"OPTIONS":      "toc:t e:t f:t pri:t todo:t tags:t",
		},
		Log: log.New(os.Stderr, "go-org: ", 0),
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
	w.Before(d)
	w.WriteNodes(d.Nodes...)
	w.After(d)
	return w.String(), err
}

func (dIn *Document) Parse(input io.Reader) (d *Document) {
	d = dIn
	defer func() {
		if recovered := recover(); recovered != nil {
			d.Error = fmt.Errorf("could not parse input: %v", recovered)
		}
	}()
	if d.tokens != nil {
		d.Error = fmt.Errorf("parse was called multiple times")
	}
	d.tokenize(input)
	_, nodes := d.parseMany(0, func(d *Document, i int) bool { return i >= len(d.tokens) })
	d.Nodes = nodes
	return d
}

func (d *Document) SetPath(path string) *Document {
	d.Path = path
	d.Log.SetPrefix(fmt.Sprintf("%s(%s): ", d.Log.Prefix(), path))
	return d
}

func (d *Document) Silent() *Document {
	d.Log = log.New(ioutil.Discard, "", 0)
	return d
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

// see https://orgmode.org/manual/Export-settings.html
func (d *Document) GetOption(key string) bool {
	get := func(settings map[string]string) string {
		for _, field := range strings.Fields(settings["OPTIONS"]) {
			if strings.HasPrefix(field, key+":") {
				return field[len(key)+1:]
			}
		}
		return ""
	}
	value := get(d.BufferSettings)
	if value == "" {
		value = get(d.DefaultSettings)
	}
	switch value {
	case "t":
		return true
	case "nil":
		return false
	default:
		d.Log.Printf("Bad value for export option %s (%s)", key, value)
		return false
	}
}

func (d *Document) parseOne(i int, stop stopFn) (consumed int, node Node) {
	switch d.tokens[i].kind {
	case "unorderedList", "orderedList":
		consumed, node = d.parseList(i, stop)
	case "tableRow", "tableSeparator":
		consumed, node = d.parseTable(i, stop)
	case "beginBlock":
		consumed, node = d.parseBlock(i, stop)
	case "beginDrawer":
		consumed, node = d.parseDrawer(i, stop)
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
	d.Log.Printf("Could not parse token %#v: Falling back to treating it as plain text.", d.tokens[i])
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

func (d *Document) addFootnote(name string, definition *FootnoteDefinition) {
	if definition != nil {
		if _, exists := d.Footnotes.Definitions[name]; exists {
			d.Log.Printf("Footnote [fn:%s] redefined! %#v", name, definition)
		}
		d.Footnotes.Definitions[name] = definition
	}
	d.Footnotes.addOrder = append(d.Footnotes.addOrder, name)
}

func (d *Document) addHeadline(headline *Headline) int {
	current := &Section{Headline: headline}
	d.Outline.last.add(current)
	d.Outline.count++
	d.Outline.last = current
	return d.Outline.count
}

func tokenize(line string) token {
	for _, lexFn := range lexFns {
		if token, ok := lexFn(line); ok {
			return token
		}
	}
	panic(fmt.Sprintf("could not lex line: %s", line))
}
