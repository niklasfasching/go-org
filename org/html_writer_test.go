package org

import (
	"strings"
	"testing"
)

type ExtendedHTMLWriter struct {
	*HTMLWriter
	callCount int
}

func (w *ExtendedHTMLWriter) WriteText(t Text) {
	w.callCount++
	w.HTMLWriter.WriteText(t)
}

func TestHTMLWriter(t *testing.T) {
	testWriter(t, func() Writer { return NewHTMLWriter() }, ".html")
}

func TestExtendedHTMLWriter(t *testing.T) {
	p := Paragraph{Children: []Node{Text{Content: "text"}, Text{Content: "more text"}}}
	htmlWriter := NewHTMLWriter()
	extendedWriter := &ExtendedHTMLWriter{htmlWriter, 0}
	htmlWriter.ExtendingWriter = extendedWriter
	WriteNodes(extendedWriter, p)
	if extendedWriter.callCount != 2 {
		t.Errorf("WriteText method of extending writer was not called: CallCount %d", extendedWriter.callCount)
	}
}

var prettyRelativeLinkTests = map[string]string{
	"[[/hello.org][hello]]": `<p><a href="/hello/">hello</a></p>`,
	"[[hello.org][hello]]":  `<p><a href="../hello/">hello</a></p>`,
	"[[file:/hello.org]]":   `<p><a href="/hello/">/hello/</a></p>`,
	"[[file:hello.org]]":    `<p><a href="../hello/">../hello/</a></p>`,
	"[[http://hello.org]]":  `<p><a href="http://hello.org">http://hello.org</a></p>`,
	"[[/foo.png]]":          `<p><img src="/foo.png" alt="/foo.png" title="/foo.png" /></p>`,
	"[[foo.png]]":           `<p><img src="../foo.png" alt="../foo.png" title="../foo.png" /></p>`,
	"[[/foo.png][foo]]":     `<p><a href="/foo.png">foo</a></p>`,
	"[[foo.png][foo]]":      `<p><a href="../foo.png">foo</a></p>`,
}

func TestPrettyRelativeLinks(t *testing.T) {
	for org, expected := range prettyRelativeLinkTests {
		t.Run(org, func(t *testing.T) {
			writer := NewHTMLWriter()
			writer.PrettyRelativeLinks = true
			actual, err := New().Silent().Parse(strings.NewReader(org), "./prettyRelativeLinkTests.org").Write(writer)
			if err != nil {
				t.Errorf("%s\n got error: %s", org, err)
			} else if actual := strings.TrimSpace(actual); actual != expected {
				t.Errorf("%s:\n%s'", org, diff(actual, expected))
			}
		})
	}
}
