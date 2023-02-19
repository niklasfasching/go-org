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

var topLevelHLevelTests = map[struct {
	TopLevelHLevel int
	input          string
}]string{
	{1, "* Top-level headline"}:        "<h1 id=\"headline-1\">\nTop-level headline\n</h1>",
	{1, "** Second-level headline"}:    "<h2 id=\"headline-1\">\nSecond-level headline\n</h2>",
	{1, "*** Third-level headline"}:    "<h3 id=\"headline-1\">\nThird-level headline\n</h3>",
	{1, "**** Fourth-level headline"}:  "<h4 id=\"headline-1\">\nFourth-level headline\n</h4>",
	{1, "***** Fifth-level headline"}:  "<h5 id=\"headline-1\">\nFifth-level headline\n</h5>",
	{1, "****** Sixth-level headline"}: "<h6 id=\"headline-1\">\nSixth-level headline\n</h6>",

	{2, "* Top-level headline"}:       "<h2 id=\"headline-1\">\nTop-level headline\n</h2>",
	{2, "** Second-level headline"}:   "<h3 id=\"headline-1\">\nSecond-level headline\n</h3>",
	{2, "*** Third-level headline"}:   "<h4 id=\"headline-1\">\nThird-level headline\n</h4>",
	{2, "**** Fourth-level headline"}: "<h5 id=\"headline-1\">\nFourth-level headline\n</h5>",
	{2, "***** Fifth-level headline"}: "<h6 id=\"headline-1\">\nFifth-level headline\n</h6>",

	{3, "* Top-level headline"}:       "<h3 id=\"headline-1\">\nTop-level headline\n</h3>",
	{3, "** Second-level headline"}:   "<h4 id=\"headline-1\">\nSecond-level headline\n</h4>",
	{3, "*** Third-level headline"}:   "<h5 id=\"headline-1\">\nThird-level headline\n</h5>",
	{3, "**** Fourth-level headline"}: "<h6 id=\"headline-1\">\nFourth-level headline\n</h6>",

	{4, "* Top-level headline"}:     "<h4 id=\"headline-1\">\nTop-level headline\n</h4>",
	{4, "** Second-level headline"}: "<h5 id=\"headline-1\">\nSecond-level headline\n</h5>",
	{4, "*** Third-level headline"}: "<h6 id=\"headline-1\">\nThird-level headline\n</h6>",

	{5, "* Top-level headline"}:     "<h5 id=\"headline-1\">\nTop-level headline\n</h5>",
	{5, "** Second-level headline"}: "<h6 id=\"headline-1\">\nSecond-level headline\n</h6>",

	{6, "* Top-level headline"}: "<h6 id=\"headline-1\">\nTop-level headline\n</h6>",
}

func TestTopLevelHLevel(t *testing.T) {
	for org, expected := range topLevelHLevelTests {
		t.Run(org.input, func(t *testing.T) {
			writer := NewHTMLWriter()
			writer.TopLevelHLevel = org.TopLevelHLevel
			actual, err := New().Silent().Parse(strings.NewReader(org.input), "./topLevelHLevelTests.org").Write(writer)
			if err != nil {
				t.Errorf("TopLevelHLevel=%d %s\n got error: %s", org.TopLevelHLevel, org.input, err)
			} else if actual := strings.TrimSpace(actual); !strings.Contains(actual, expected) {
				t.Errorf("TopLevelHLevel=%d %s:\n%s'", org.TopLevelHLevel, org.input, diff(actual, expected))
			}
		})
	}
}
