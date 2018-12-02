package org

import (
	"fmt"
	"html"
	"path"
	"strings"
)

type HTMLWriter struct {
	stringBuilder
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
	"^{}": []string{"<super>", "</super>"},
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
		HighlightCodeBlock: func(source, lang string) string { return html.EscapeString(source) },
	}
}

func (w *HTMLWriter) emptyClone() *HTMLWriter {
	wcopy := *w
	wcopy.stringBuilder = stringBuilder{}
	return &wcopy
}

func (w *HTMLWriter) before(d *Document) {}

func (w *HTMLWriter) after(d *Document) {
	fs := d.Footnotes
	if len(fs.Definitions) == 0 {
		return
	}
	w.WriteString(`<div id="footnotes">` + "\n")
	w.WriteString(`<h1 class="footnotes-title">` + fs.Title + `</h1>` + "\n")
	w.WriteString(`<div class="footnote-definitions">` + "\n")
	for _, name := range fs.Order {
		w.writeNodes(fs.Definitions[name])
	}
	w.WriteString("</div>\n</div>\n")
}

func (w *HTMLWriter) writeNodes(ns ...Node) {
	for _, n := range ns {
		switch n := n.(type) {
		case Keyword, Comment:
			continue
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
		case Line:
			w.writeLine(n)

		case Text:
			w.writeText(n)
		case Emphasis:
			w.writeEmphasis(n)
		case Linebreak:
			w.writeLinebreak(n)
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

func (w *HTMLWriter) writeLines(lines []Node) {
	for i, line := range lines {
		w.writeNodes(line)
		if i != len(lines)-1 && line.(Line).Children != nil {
			w.WriteString(" ")
		}
	}
}

func (w *HTMLWriter) writeBlock(b Block) {
	w.WriteString("<code>")
	lang := ""
	if len(b.Parameters) >= 1 {
		lang = b.Parameters[0]
	}
	lines := []string{}
	for _, n := range b.Children {
		lines = append(lines, n.(Line).Children[0].(Text).Content)
	}
	w.WriteString(w.HighlightCodeBlock(strings.Join(lines, "\n"), lang))
	w.WriteString("</code>\n")
}

func (w *HTMLWriter) writeFootnoteDefinition(f FootnoteDefinition) {
	w.WriteString(`<div class="footnote-definition">` + "\n")
	w.WriteString(fmt.Sprintf(`<sup id="footnote-%s">%s</sup>`, f.Name, f.Name) + "\n")
	w.writeNodes(f.Children...)
	w.WriteString("</div>\n")
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

func (w *HTMLWriter) writeLinebreak(l Linebreak) {
	w.WriteString("<br>\n")
}

func (w *HTMLWriter) writeFootnoteLink(l FootnoteLink) {
	name := html.EscapeString(l.Name)
	w.WriteString(fmt.Sprintf(`<sup class="footnote-reference"><a href="#footnote-%s">%s</a></sup>`, name, name))

}

func (w *HTMLWriter) writeRegularLink(l RegularLink) {
	url := html.EscapeString(l.URL)
	descriptionWriter := w.emptyClone()
	descriptionWriter.writeNodes(l.Description...)
	description := descriptionWriter.String()
	switch l.Protocol {
	case "file": // TODO
		url = url[len("file:"):]
		if strings.Contains(".png.jpg.jpeg.gif", path.Ext(l.URL)) {
			w.WriteString(fmt.Sprintf(`<img src="%s" alt="%s" title="%s" />`, url, description, description))
		} else {
			w.WriteString(fmt.Sprintf(`<a href="%s">%s</a>`, url, description))
		}
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
	w.WriteString("<li>")
	if len(li.Children) == 1 {
		if p, ok := li.Children[0].(Paragraph); ok {
			w.writeLines(p.Children)
		}
	} else {
		w.writeNodes(li.Children...)
	}
	w.WriteString("</li>\n")
}

func (w *HTMLWriter) writeLine(l Line) {
	w.writeNodes(l.Children...)
}

func (w *HTMLWriter) writeParagraph(p Paragraph) {
	if len(p.Children) == 1 && p.Children[0].(Line).Children == nil {
		return
	}
	w.WriteString("<p>")
	w.writeLines(p.Children)
	w.WriteString("</p>\n")
}

func (w *HTMLWriter) writeHorizontalRule(h HorizontalRule) {
	w.WriteString("<hr>\n")
}

func (w *HTMLWriter) writeTable(t Table) {
	w.WriteString("<table>")
	w.writeNodes(t.Header)
	w.WriteString("<tbody>")
	w.writeNodes(t.Rows...)
	w.WriteString("</tbody>\n</table>\n")
}

func (w *HTMLWriter) writeTableRow(t TableRow) {
	w.WriteString("\n<tr>\n")
	for _, column := range t.Columns {
		w.WriteString("<td>")
		w.writeNodes(column...)
		w.WriteString("</td>")
	}
	w.WriteString("\n</tr>\n")
}

func (w *HTMLWriter) writeTableHeader(t TableHeader) {
	w.WriteString("\n<thead>\n")
	for _, column := range t.Columns {
		w.WriteString("<th>")
		w.writeNodes(column...)
		w.WriteString("</th>")
	}
	w.WriteString("\n</thead>\n")
}

func (w *HTMLWriter) writeTableSeparator(t TableSeparator) {
	w.WriteString("\n<tr></tr>\n")
}
