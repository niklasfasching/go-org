package org

import (
	"fmt"
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
func (w *OrgWriter) after(d *Document)  {}

func (w *OrgWriter) emptyClone() *OrgWriter {
	wcopy := *w
	wcopy.stringBuilder = strings.Builder{}
	return &wcopy
}

func (w *OrgWriter) nodesAsString(nodes ...Node) string {
	tmp := w.emptyClone()
	tmp.writeNodes(nodes...)
	return tmp.String()
}

func (w *OrgWriter) writeNodes(ns ...Node) {
	for _, n := range ns {
		switch n := n.(type) {
		case Comment:
			w.writeComment(n)
		case Keyword:
			w.writeKeyword(n)
		case Include:
			w.writeKeyword(n.Keyword)
		case NodeWithMeta:
			w.writeNodeWithMeta(n)
		case Headline:
			w.writeHeadline(n)
		case Block:
			w.writeBlock(n)
		case Drawer:
			w.writeDrawer(n)

		case FootnoteDefinition:
			w.writeFootnoteDefinition(n)

		case List:
			w.writeList(n)
		case ListItem:
			w.writeListItem(n)
		case DescriptiveListItem:
			w.writeDescriptiveListItem(n)

		case Table:
			w.writeTable(n)

		case Paragraph:
			w.writeParagraph(n)
		case Example:
			w.writeExample(n)
		case HorizontalRule:
			w.writeHorizontalRule(n)
		case Text:
			w.writeText(n)
		case Emphasis:
			w.writeEmphasis(n)
		case LineBreak:
			w.writeLineBreak(n)
		case ExplicitLineBreak:
			w.writeExplicitLineBreak(n)
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
	if h.Properties != nil {
		w.writeNodes(h.Properties)
	}
	w.writeNodes(h.Children...)
}

func (w *OrgWriter) writeBlock(b Block) {
	w.WriteString(w.indent + "#+BEGIN_" + b.Name)
	if len(b.Parameters) != 0 {
		w.WriteString(" " + strings.Join(b.Parameters, " "))
	}
	w.WriteString("\n" + w.indent)
	w.writeNodes(b.Children...)
	w.WriteString("#+END_" + b.Name + "\n")
}

func (w *OrgWriter) writeDrawer(d Drawer) {
	w.WriteString(w.indent + ":" + d.Name + ":\n")
	w.writeNodes(d.Children...)
	w.WriteString(w.indent + ":END:\n")
}

func (w *OrgWriter) writeFootnoteDefinition(f FootnoteDefinition) {
	w.WriteString(fmt.Sprintf("[fn:%s]", f.Name))
	if !(len(f.Children) >= 1 && isEmptyLineParagraph(f.Children[0])) {
		w.WriteString(" ")
	}
	w.writeNodes(f.Children...)
}

func (w *OrgWriter) writeParagraph(p Paragraph) {
	w.writeNodes(p.Children...)
	w.WriteString("\n")
}

func (w *OrgWriter) writeExample(e Example) {
	for _, n := range e.Children {
		w.WriteString(w.indent + ":" + " ")
		w.writeNodes(n)
		w.WriteString("\n")
	}
}

func (w *OrgWriter) writeKeyword(k Keyword) {
	w.WriteString(w.indent + fmt.Sprintf("#+%s: %s\n", k.Key, k.Value))
}

func (w *OrgWriter) writeNodeWithMeta(n NodeWithMeta) {
	for _, ns := range n.Meta.Caption {
		w.WriteString("#+CAPTION: ")
		w.writeNodes(ns...)
		w.WriteString("\n")
	}
	for _, attributes := range n.Meta.HTMLAttributes {
		w.WriteString("#+ATTR_HTML: ")
		w.WriteString(strings.Join(attributes, " ") + "\n")
	}
	w.writeNodes(n.Node)
}

func (w *OrgWriter) writeComment(c Comment) {
	w.WriteString(w.indent + "#" + c.Content)
}

func (w *OrgWriter) writeList(l List) { w.writeNodes(l.Items...) }

func (w *OrgWriter) writeListItem(li ListItem) {
	liWriter := w.emptyClone()
	liWriter.indent = w.indent + strings.Repeat(" ", len(li.Bullet)+1)
	liWriter.writeNodes(li.Children...)
	content := strings.TrimPrefix(liWriter.String(), liWriter.indent)
	w.WriteString(w.indent + li.Bullet)
	if len(content) > 0 && content[0] == '\n' {
		w.WriteString(content)
	} else {
		w.WriteString(" " + content)
	}
}

func (w *OrgWriter) writeDescriptiveListItem(di DescriptiveListItem) {
	w.WriteString(w.indent + di.Bullet)
	indent := w.indent + strings.Repeat(" ", len(di.Bullet)+1)
	if len(di.Term) != 0 {
		term := w.nodesAsString(di.Term...)
		w.WriteString(" " + term + " ::")
		indent = indent + strings.Repeat(" ", len(term)+4)
	}
	diWriter := w.emptyClone()
	diWriter.indent = indent
	diWriter.writeNodes(di.Details...)
	details := strings.TrimPrefix(diWriter.String(), diWriter.indent)
	if len(details) > 0 && details[0] == '\n' {
		w.WriteString(details)
	} else {
		w.WriteString(" " + details)
	}
}

func (w *OrgWriter) writeTable(t Table) {
	for _, row := range t.Rows {
		w.WriteString(w.indent)
		if len(row.Columns) == 0 {
			w.WriteString(`|`)
			for i := 0; i < len(t.ColumnInfos); i++ {
				w.WriteString(strings.Repeat("-", t.ColumnInfos[i].Len+2))
				if i < len(t.ColumnInfos)-1 {
					w.WriteString("+")
				}
			}
			w.WriteString(`|`)

		} else {
			w.WriteString(`|`)
			for _, column := range row.Columns {
				w.WriteString(` `)
				content := w.nodesAsString(column.Children...)
				if content == "" {
					content = " "
				}
				n := column.Len - len(content)
				if n < 0 {
					n = 0
				}
				if column.Align == "center" {
					if n%2 != 0 {
						w.WriteString(" ")
					}
					w.WriteString(strings.Repeat(" ", n/2) + content + strings.Repeat(" ", n/2))
				} else if column.Align == "right" {
					w.WriteString(strings.Repeat(" ", n) + content)
				} else {
					w.WriteString(content + strings.Repeat(" ", n))
				}
				w.WriteString(` |`)
			}
		}
		w.WriteString("\n")
	}
}

func (w *OrgWriter) writeHorizontalRule(hr HorizontalRule) {
	w.WriteString(w.indent + "-----\n")
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

func (w *OrgWriter) writeLineBreak(l LineBreak) {
	w.WriteString(strings.Repeat("\n"+w.indent, l.Count))
}

func (w *OrgWriter) writeExplicitLineBreak(l ExplicitLineBreak) {
	w.WriteString(`\\` + "\n" + w.indent)
}

func (w *OrgWriter) writeFootnoteLink(l FootnoteLink) {
	w.WriteString("[fn:" + l.Name)
	if l.Definition != nil {
		w.WriteString(":")
		w.writeNodes(l.Definition.Children[0].(Paragraph).Children...)
	}
	w.WriteString("]")
}

func (w *OrgWriter) writeRegularLink(l RegularLink) {
	if l.AutoLink {
		w.WriteString(l.URL)
	} else if l.Description == nil {
		w.WriteString(fmt.Sprintf("[[%s]]", l.URL))
	} else {
		descriptionWriter := w.emptyClone()
		descriptionWriter.writeNodes(l.Description...)
		description := descriptionWriter.String()
		w.WriteString(fmt.Sprintf("[[%s][%s]]", l.URL, description))
	}
}
