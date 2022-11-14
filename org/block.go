package org

import (
	"math"
	"regexp"
	"strings"
	"unicode"
)

type Block struct {
	Name       string
	Parameters []string
	Children   []RangedNode
	Result     RangedNode
}

type Result struct {
	Node RangedNode
}

type Example struct {
	Children []RangedNode
}

type LatexBlock struct {
	Content []Node
}

var exampleLineRegexp = regexp.MustCompile(`^(\s*):(\s(.*)|\s*$)`)
var beginBlockRegexp = regexp.MustCompile(`(?i)^(\s*)#\+BEGIN_(\w+)(.*)`)
var endBlockRegexp = regexp.MustCompile(`(?i)^(\s*)#\+END_(\w+)`)
var beginLatexBlockRegexp = regexp.MustCompile(`(?i)^(\s*)\\begin{([^}]+)}(\s*)$`)
var endLatexBlockRegexp = regexp.MustCompile(`(?i)^(\s*)\\end{([^}]+)}(\s*)$`)
var resultRegexp = regexp.MustCompile(`(?i)^(\s*)#\+RESULTS:`)
var exampleBlockEscapeRegexp = regexp.MustCompile(`(^|\n)([ \t]*),([ \t]*)(\*|,\*|#\+|,#\+)`)

func lexBlock(line string) (token, bool) {
	if m := beginBlockRegexp.FindStringSubmatch(line); m != nil {
		return token{"beginBlock", len(m[1]), strings.ToUpper(m[2]), m}, true
	} else if m := endBlockRegexp.FindStringSubmatch(line); m != nil {
		return token{"endBlock", len(m[1]), strings.ToUpper(m[2]), m}, true
	}
	return nilToken, false
}

func lexLatexBlock(line string) (token, bool) {
	if m := beginLatexBlockRegexp.FindStringSubmatch(line); m != nil {
		return token{"beginLatexBlock", len(m[1]), strings.ToUpper(m[2]), m}, true
	} else if m := endLatexBlockRegexp.FindStringSubmatch(line); m != nil {
		return token{"endLatexBlock", len(m[1]), strings.ToUpper(m[2]), m}, true
	}
	return nilToken, false
}

func lexResult(line string) (token, bool) {
	if m := resultRegexp.FindStringSubmatch(line); m != nil {
		return token{"result", len(m[1]), "", m}, true
	}
	return nilToken, false
}

func lexExample(line string) (token, bool) {
	if m := exampleLineRegexp.FindStringSubmatch(line); m != nil {
		return token{"example", len(m[1]), m[3], m}, true
	}
	return nilToken, false
}

func isRawTextBlock(name string) bool { return name == "SRC" || name == "EXAMPLE" || name == "EXPORT" }

func (d *Document) parseBlock(i int, parentStop stopFn) (int, RangedNode) {
	t, start := d.tokens[i], i
	name, parameters := t.content, splitParameters(t.matches[3])
	trim := trimIndentUpTo(d.tokens[i].lvl)
	stop := func(d *Document, i int) bool {
		return i >= len(d.tokens) || (d.tokens[i].kind == "endBlock" && d.tokens[i].content == name)
	}
	block, i := Block{name, parameters, nil, RangedNode{}}, i+1
	if isRawTextBlock(name) {
		rawText := ""
		for ; !stop(d, i); i++ {
			rawText += trim(d.tokens[i].matches[0]) + "\n"
		}
		if name == "EXAMPLE" || (name == "SRC" && len(parameters) >= 1 && parameters[0] == "org") {
			rawText = exampleBlockEscapeRegexp.ReplaceAllString(rawText, "$1$2$3$4")
		}
		block.Children = d.parseRawInline(rawText)
	} else {
		consumed, nodes := d.parseMany(i, stop)
		block.Children = nodes
		i += consumed
	}
	if i >= len(d.tokens) || d.tokens[i].kind != "endBlock" || d.tokens[i].content != name {
		return 0, RangedNode{}
	}
	if name == "SRC" {
		consumed, result := d.parseSrcBlockResult(i+1, parentStop)
		block.Result = result
		i += consumed
	}
	return i + 1 - start, RangedNode{block, start, i + 1}
}

func (d *Document) parseLatexBlock(i int, parentStop stopFn) (int, Node) {
	t, start := d.tokens[i], i
	name, rawText, trim := t.content, "", trimIndentUpTo(int(math.Max((float64(d.baseLvl)), float64(t.lvl))))
	stop := func(d *Document, i int) bool {
		return i >= len(d.tokens) || (d.tokens[i].kind == "endLatexBlock" && d.tokens[i].content == name)
	}
	for ; !stop(d, i); i++ {
		rawText += trim(d.tokens[i].matches[0]) + "\n"
	}
	if i >= len(d.tokens) || d.tokens[i].kind != "endLatexBlock" || d.tokens[i].content != name {
		return 0, nil
	}
	rawText += trim(d.tokens[i].matches[0])
	return i + 1 - start, LatexBlock{d.parseRawInline(rawText)}
}

func (d *Document) parseSrcBlockResult(i int, parentStop stopFn) (int, RangedNode) {
	start := i
	for ; !parentStop(d, i) && d.tokens[i].kind == "text" && d.tokens[i].content == ""; i++ {
	}
	if parentStop(d, i) || d.tokens[i].kind != "result" {
		return 0, RangedNode{}
	}
	consumed, result := d.parseResult(i, parentStop)
	return (i - start) + consumed, result
}

func (d *Document) parseExample(i int, parentStop stopFn) (int, RangedNode) {
	example, start := Example{}, i
	for ; !parentStop(d, i) && d.tokens[i].kind == "example"; i++ {
		example.Children = append(example.Children, RangedNode{Text{d.tokens[i].content, true}, start, i})
	}
	return i - start, RangedNode{example, start, i}
}

func (d *Document) parseResult(i int, parentStop stopFn) (int, RangedNode) {
	if i+1 >= len(d.tokens) {
		return 0, RangedNode{}
	}
	consumed, node := d.parseOne(i+1, parentStop)
	return consumed + 1, RangedNode{Result{node}, i, consumed + 1}
}

func trimIndentUpTo(max int) func(string) string {
	return func(line string) string {
		i := 0
		for ; i < len(line) && i < max && unicode.IsSpace(rune(line[i])); i++ {
		}
		return line[i:]
	}
}

func splitParameters(s string) []string {
	parameters, parts := []string{}, strings.Split(s, " :")
	lang, rest := strings.TrimSpace(parts[0]), parts[1:]
	if lang != "" {
		parameters = append(parameters, lang)
	}
	for _, p := range rest {
		kv := strings.SplitN(p+" ", " ", 2)
		parameters = append(parameters, ":"+kv[0], strings.TrimSpace(kv[1]))
	}
	return parameters
}

func (b Block) ParameterMap() map[string]string {
	if len(b.Parameters) == 0 {
		return nil
	}
	m := map[string]string{":lang": b.Parameters[0]}
	for i := 1; i+1 < len(b.Parameters); i += 2 {
		m[b.Parameters[i]] = b.Parameters[i+1]
	}
	return m
}

func (n Example) String() string    { return orgWriter.WriteNodesAsString(n) }
func (n Block) String() string      { return orgWriter.WriteNodesAsString(n) }
func (n LatexBlock) String() string { return orgWriter.WriteNodesAsString(n) }
func (n Result) String() string     { return orgWriter.WriteNodesAsString(n) }
