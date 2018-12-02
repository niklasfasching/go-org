package org

import (
	"fmt"
	"regexp"
	"strings"
)

type stringBuilder = strings.Builder

type OrgWriter struct {
	TagsColumn int // see org-tags-column
	stringBuilder
	indent string
}

var emphasisOrgBorders = map[string][]string{
	"_":   []string{"_", "_"},
	"*":   []string{"*", "*"},
	"/":   []string{"/", "/"},
	"+":   []string{"+", "+"},
	"~":   []string{"~", "~"},
	"=":   []string{"=", "="},
	"_{}": []string{"_{", "}"},
	"^{}": []string{"^{", "}"},
}

func NewOrgWriter() *OrgWriter {
	return &OrgWriter{
		TagsColumn: 77,
	}
}

func (w *OrgWriter) before(d *Document) {}
func (w *OrgWriter) after(d *Document) {
	w.writeFootnotes(d)
}

func (w *OrgWriter) emptyClone() *OrgWriter {
	wcopy := *w
	wcopy.stringBuilder = strings.Builder{}
	return &wcopy
}

func (w *OrgWriter) writeNodes(ns ...Node) {
	for _, n := range ns {
		switch n := n.(type) {
		case Comment:
			w.writeComment(n)
		case Keyword:
			w.writeKeyword(n)
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

var eolWhiteSpaceRegexp = regexp.MustCompile("[\t ]*\n")

func (w *OrgWriter) String() string {
	s := w.stringBuilder.String()
	return eolWhiteSpaceRegexp.ReplaceAllString(s, "\n")
}

func (w *OrgWriter) writeHeadline(h Headline) {
	tmp := w.emptyClone()
	tmp.WriteString(strings.Repeat("*", h.Lvl))
	if h.Status != "" {
		tmp.WriteString(" " + h.Status)
	}
	if h.Priority != "" {
		tmp.WriteString(" [#" + h.Priority + "]")
	}
	tmp.WriteString(" ")
	tmp.writeNodes(h.Title...)
	hString := tmp.String()
	if len(h.Tags) != 0 {
		hString += " "
		tString := ":" + strings.Join(h.Tags, ":") + ":"
		if n := w.TagsColumn - len(tString) - len(hString); n > 0 {
			w.WriteString(hString + strings.Repeat(" ", n) + tString)
		} else {
			w.WriteString(hString + tString)
		}
	} else {
		w.WriteString(hString)
	}
	w.WriteString("\n")
	if len(h.Children) != 0 {
		w.WriteString(w.indent)
	}
	w.writeNodes(h.Children...)
}

func (w *OrgWriter) writeBlock(b Block) {
	w.WriteString(fmt.Sprintf("%s#+BEGIN_%s %s\n", w.indent, b.Name, strings.Join(b.Parameters, " ")))
	w.writeNodes(b.Children...)
	w.WriteString(w.indent + "#+END_" + b.Name + "\n")
}

func (w *OrgWriter) writeFootnotes(d *Document) {
	fs := d.Footnotes
	if len(fs.Definitions) == 0 {
		return
	}
	w.WriteString("* " + fs.Title + "\n")
	for _, name := range fs.Order {
		if fnDefinition := fs.Definitions[name]; !fnDefinition.Inline {
			w.writeNodes(fnDefinition)
		}
	}
}

func (w *OrgWriter) writeFootnoteDefinition(f FootnoteDefinition) {
	w.WriteString(fmt.Sprintf("[fn:%s] ", f.Name))
	w.writeNodes(f.Children...)
}

func (w *OrgWriter) writeParagraph(p Paragraph) {
	w.writeNodes(p.Children...)
}

func (w *OrgWriter) writeKeyword(k Keyword) {
	w.WriteString(w.indent + fmt.Sprintf("#+%s: %s\n", k.Key, k.Value))
}

func (w *OrgWriter) writeComment(c Comment) {
	w.WriteString(w.indent + "#" + c.Content)
}

func (w *OrgWriter) writeList(l List) { w.writeNodes(l.Items...) }

func (w *OrgWriter) writeListItem(li ListItem) {
	w.WriteString(w.indent + li.Bullet + " ")
	liWriter := w.emptyClone()
	liWriter.indent = w.indent + strings.Repeat(" ", len(li.Bullet)+1)
	liWriter.writeNodes(li.Children...)
	w.WriteString(strings.TrimPrefix(liWriter.String(), liWriter.indent))
}

func (w *OrgWriter) writeTable(t Table) {
	// TODO: pretty print tables
	w.writeNodes(t.Header)
	w.writeNodes(t.Rows...)
}

func (w *OrgWriter) writeTableHeader(th TableHeader) {
	w.writeTableColumns(th.Columns)
	w.writeNodes(th.Separator)
}

func (w *OrgWriter) writeTableRow(tr TableRow) {
	w.writeTableColumns(tr.Columns)
}

func (w *OrgWriter) writeTableSeparator(ts TableSeparator) {
	w.WriteString(w.indent + ts.Content + "\n")
}

func (w *OrgWriter) writeTableColumns(columns [][]Node) {
	w.WriteString(w.indent + "| ")
	for _, columnNodes := range columns {
		w.writeNodes(columnNodes...)
		w.WriteString(" | ")
	}
	w.WriteString("\n")
}

func (w *OrgWriter) writeHorizontalRule(hr HorizontalRule) {
	w.WriteString(w.indent + "-----\n")
}

func (w *OrgWriter) writeLine(l Line) {
	w.WriteString(w.indent)
	w.writeNodes(l.Children...)
	w.WriteString("\n")
}

func (w *OrgWriter) writeText(t Text) { w.WriteString(t.Content) }

func (w *OrgWriter) writeEmphasis(e Emphasis) {
	borders, ok := emphasisOrgBorders[e.Kind]
	if !ok {
		panic(fmt.Sprintf("bad emphasis %#v", e))
	}
	w.WriteString(borders[0])
	w.writeNodes(e.Content...)
	w.WriteString(borders[1])
}

func (w *OrgWriter) writeLinebreak(l Linebreak) {
	w.WriteString(`\\`)
}

func (w *OrgWriter) writeFootnoteLink(l FootnoteLink) {
	w.WriteString("[fn:" + l.Name)
	if l.Definition != nil {
		w.WriteString(":")
		w.writeNodes(l.Definition.Children...)
	}
	w.WriteString("]")
}

func (w *OrgWriter) writeRegularLink(l RegularLink) {
	descriptionWriter := w.emptyClone()
	descriptionWriter.writeNodes(l.Description...)
	description := descriptionWriter.String()
	if l.URL != description {
		w.WriteString(fmt.Sprintf("[[%s][%s]]", l.URL, description))
	} else {
		w.WriteString(fmt.Sprintf("[[%s]]", l.URL))
	}
}
