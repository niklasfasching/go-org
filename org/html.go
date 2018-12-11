package org

import (
	"fmt"
	"html"
	"strings"
)

type HTMLWriter struct {
	stringBuilder
	document           *Document
	HighlightCodeBlock func(source, lang string) string
}

var emphasisTags = map[string][]string{
	"/":   []string{"<em>", "</em>"},
	"*":   []string{"<strong>", "</strong>"},
	"+":   []string{"<del>", "</del>"},
	"~":   []string{"<code>", "</code>"},
	"=":   []string{`<code class="verbatim">`, "</code>"},
	"_":   []string{`<span style="text-decoration: underline;">`, "</span>"},
	"_{}": []string{"<sub>", "</sub>"},
	"^{}": []string{"<sup>", "</sup>"},
}

var listTags = map[string][]string{
	"+":      []string{"<ul>", "</ul>"},
	"-":      []string{"<ul>", "</ul>"},
	"*":      []string{"<ul>", "</ul>"},
	"number": []string{"<ol>", "</ol>"},
	"letter": []string{"<ol>", "</ol>"},
}

func NewHTMLWriter() *HTMLWriter {
	return &HTMLWriter{
		HighlightCodeBlock: func(source, lang string) string {
			return fmt.Sprintf("%s\n<pre>\n%s\n</pre>\n</div>", `<div class="highlight">`, html.EscapeString(source))
		},
	}
}

func (w *HTMLWriter) emptyClone() *HTMLWriter {
	wcopy := *w
	wcopy.stringBuilder = stringBuilder{}
	return &wcopy
}

func (w *HTMLWriter) before(d *Document) {
	w.document = d
}

func (w *HTMLWriter) after(d *Document) {
	w.writeFootnotes(d)
}

func (w *HTMLWriter) writeNodes(ns ...Node) {
	for _, n := range ns {
		switch n := n.(type) {
		case Keyword, Comment:
			continue
		case NodeWithMeta:
			w.writeNodeWithMeta(n)
		case Headline:
			w.writeHeadline(n)
		case Block:
			w.writeBlock(n)

		case FootnoteDefinition:
			w.writeFootnoteDefinition(n)

		case List:
			w.writeList(n)
		case ListItem:
			w.writeListItem(n)

		case Table:
			w.writeTable(n)
		case TableHeader:
			w.writeTableHeader(n)
		case TableRow:
			w.writeTableRow(n)
		case TableSeparator:
			w.writeTableSeparator(n)

		case Paragraph:
			w.writeParagraph(n)
		case HorizontalRule:
			w.writeHorizontalRule(n)
		case Text:
			w.writeText(n)
		case Emphasis:
			w.writeEmphasis(n)
		case ExplicitLineBreak:
			w.writeExplicitLineBreak(n)
		case LineBreak:
			w.writeLineBreak(n)
		case RegularLink:
			w.writeRegularLink(n)
		case FootnoteLink:
			w.writeFootnoteLink(n)
		default:
			if n != nil {
				panic(fmt.Sprintf("bad node %#v", n))
			}
		}
	}
}

func (w *HTMLWriter) writeBlock(b Block) {
	switch name := b.Name; {
	case name == "SRC":
		source, lang := b.Children[0].(Text).Content, "text"
		if len(b.Parameters) >= 1 {
			lang = strings.ToLower(b.Parameters[0])
		}
		w.WriteString(w.HighlightCodeBlock(source, lang) + "\n")
	case name == "EXAMPLE":
		w.WriteString(`<pre class="example">` + "\n")
		w.writeNodes(b.Children...)
		w.WriteString("\n</pre>\n")
	case name == "EXPORT" && len(b.Parameters) >= 1 && strings.ToLower(b.Parameters[0]) == "html":
		w.WriteString(b.Children[0].(Text).Content + "\n")
	case name == "QUOTE":
		w.WriteString("<blockquote>\n")
		w.writeNodes(b.Children...)
		w.WriteString("</blockquote>\n")
	case name == "CENTER":
		w.WriteString(`<div class="center-block" style="text-align: center; margin-left: auto; margin-right: auto;">` + "\n")
		w.writeNodes(b.Children...)
		w.WriteString("</div>\n")
	default:
		w.WriteString(fmt.Sprintf(`<div class="%s-block">`, strings.ToLower(b.Name)) + "\n")
		w.writeNodes(b.Children...)
		w.WriteString("</div>\n")
	}
}

func (w *HTMLWriter) writeFootnoteDefinition(f FootnoteDefinition) {
	n := f.Name
	w.WriteString(`<div class="footnote-definition">` + "\n")
	w.WriteString(fmt.Sprintf(`<sup id="footnote-%s"><a href="#footnote-reference-%s">%s</a></sup>`, n, n, n) + "\n")
	w.WriteString(`<div class="footnote-body">` + "\n")
	w.writeNodes(f.Children...)
	w.WriteString("</div>\n</div>\n")
}

func (w *HTMLWriter) writeFootnotes(d *Document) {
	fs := d.Footnotes
	if len(fs.Definitions) == 0 {
		return
	}
	w.WriteString(`<div class="footnotes">` + "\n")
	w.WriteString(`<h1 class="footnotes-title">` + fs.Title + `</h1>` + "\n")
	w.WriteString(`<div class="footnote-definitions">` + "\n")
	for _, definition := range d.Footnotes.Ordered() {
		w.writeNodes(definition)
	}
	w.WriteString("</div>\n</div>\n")
}

func (w *HTMLWriter) writeHeadline(h Headline) {
	w.WriteString(fmt.Sprintf("<h%d>", h.Lvl))
	w.writeNodes(h.Title...)
	w.WriteString(fmt.Sprintf("</h%d>\n", h.Lvl))
	w.writeNodes(h.Children...)
}

func (w *HTMLWriter) writeText(t Text) {
	w.WriteString(html.EscapeString(t.Content))
}

func (w *HTMLWriter) writeEmphasis(e Emphasis) {
	tags, ok := emphasisTags[e.Kind]
	if !ok {
		panic(fmt.Sprintf("bad emphasis %#v", e))
	}
	w.WriteString(tags[0])
	w.writeNodes(e.Content...)
	w.WriteString(tags[1])
}

func (w *HTMLWriter) writeLineBreak(l LineBreak) {
	w.WriteString("\n")
}

func (w *HTMLWriter) writeExplicitLineBreak(l ExplicitLineBreak) {
	w.WriteString("<br>\n")
}

func (w *HTMLWriter) writeFootnoteLink(l FootnoteLink) {
	n := html.EscapeString(l.Name)
	w.WriteString(fmt.Sprintf(`<sup class="footnote-reference"><a id="footnote-reference-%s" href="#footnote-%s">%s</a></sup>`, n, n, n))
}

func (w *HTMLWriter) writeRegularLink(l RegularLink) {
	url := html.EscapeString(l.URL)
	if l.Protocol == "file" {
		url = url[len("file:"):]
	}
	description := url
	if l.Description != nil {
		descriptionWriter := w.emptyClone()
		descriptionWriter.writeNodes(l.Description...)
		description = descriptionWriter.String()
	}
	switch l.Kind() {
	case "image":
		w.WriteString(fmt.Sprintf(`<img src="%s" alt="%s" title="%s" />`, url, description, description))
	case "video":
		w.WriteString(fmt.Sprintf(`<video src="%s" title="%s">%s</video>`, url, description, description))
	default:
		w.WriteString(fmt.Sprintf(`<a href="%s">%s</a>`, url, description))
	}
}

func (w *HTMLWriter) writeList(l List) {
	tags, ok := listTags[l.Kind]
	if !ok {
		panic(fmt.Sprintf("bad list kind %#v", l))
	}
	w.WriteString(tags[0] + "\n")
	w.writeNodes(l.Items...)
	w.WriteString(tags[1] + "\n")
}

func (w *HTMLWriter) writeListItem(li ListItem) {
	w.WriteString("<li>\n")
	w.writeNodes(li.Children...)
	w.WriteString("</li>\n")
}

func (w *HTMLWriter) writeParagraph(p Paragraph) {
	if isEmptyLineParagraph(p) {
		return
	}
	w.WriteString("<p>")
	if _, ok := p.Children[0].(LineBreak); !ok {
		w.WriteString("\n")
	}
	w.writeNodes(p.Children...)
	w.WriteString("\n</p>\n")
}

func (w *HTMLWriter) writeHorizontalRule(h HorizontalRule) {
	w.WriteString("<hr>\n")
}

func (w *HTMLWriter) writeNodeWithMeta(m NodeWithMeta) {
	nodeW := w.emptyClone()
	nodeW.writeNodes(m.Node)
	nodeString := nodeW.String()
	if rawCaption, ok := m.Meta["CAPTION"]; ok {
		nodes, captionW := w.document.parseInline(rawCaption), w.emptyClone()
		captionW.writeNodes(nodes...)
		caption := `<p class="caption">` + "\n" + captionW.String() + "\n</p>\n"
		nodeString = `<div class="captioned">` + "\n" + nodeString + caption + `</div>` + "\n"
	}
	w.WriteString(nodeString)
}

func (w *HTMLWriter) writeTable(t Table) {
	w.WriteString("<table>\n")
	w.writeNodes(t.Header)
	w.WriteString("<tbody>\n")
	w.writeNodes(t.Rows...)
	w.WriteString("</tbody>\n</table>\n")
}

func (w *HTMLWriter) writeTableRow(t TableRow) {
	w.WriteString("<tr>\n")
	for _, column := range t.Columns {
		w.WriteString("<td>")
		w.writeNodes(column...)
		w.WriteString("</td>")
	}
	w.WriteString("\n</tr>\n")
}

func (w *HTMLWriter) writeTableHeader(t TableHeader) {
	w.WriteString("<thead>\n")
	for _, column := range t.Columns {
		w.WriteString("<th>")
		w.writeNodes(column...)
		w.WriteString("</th>")
	}
	w.WriteString("\n</thead>\n")
}

func (w *HTMLWriter) writeTableSeparator(t TableSeparator) {
	w.WriteString("<tr></tr>\n")
}
