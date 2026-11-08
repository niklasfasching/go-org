package org

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
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

func TestHTMLMacroParagraph(t *testing.T) {
	tests := []struct {
		name       string
		definition string
		input      string
		want       string
	}{
		{"inline", "Hello", "xxx {{{test()}}} yyy", "<p>xxx Hello yyy</p>"},
		{"first", "Hello", "{{{test()}}} yyy", "<p>Hello yyy</p>"},
		{"last", "Hello", "xxx {{{test()}}}", "<p>xxx Hello</p>"},
		{"standalone", "Hello", "{{{test()}}}", "<p>Hello</p>"},
		{"parameters", "newline added after $1 and $2 inserted", "How to avoid {{{test(foo, bar)}}} ?", "<p>How to avoid newline added after foo and  bar inserted ?</p>"},
		{"expansion markup", "*Hello* [[https://example.test/][link]]", "xxx {{{test()}}} yyy", `<p>xxx <strong>Hello</strong> <a href="https://example.test/">link</a> yyy</p>`},
		{"surrounding markup", "Hello", "*xxx* {{{test()}}} /yyy/", "<p><strong>xxx</strong> Hello <em>yyy</em></p>"},
		{"undefined", "Hello", "xxx {{{missing()}}} yyy", "<p>xxx  yyy</p>"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := "#+MACRO: test " + test.definition + "\n" + test.input
			document := New().Silent().Parse(strings.NewReader(input), "macro.org")
			actual, err := document.Write(NewHTMLWriter())
			if err != nil {
				t.Fatal(err)
			}
			if strings.TrimSpace(actual) != test.want {
				t.Fatalf("unexpected HTML:\n%s", diff(actual, test.want+"\n"))
			}
			// A browser repairs nested <p> elements, so assert the raw output
			// above as well as the resulting paragraph structure.
			root, err := html.Parse(strings.NewReader(actual))
			if err != nil {
				t.Fatal(err)
			}
			paragraphs := 0
			var visit func(*html.Node)
			visit = func(node *html.Node) {
				if node.Type == html.ElementNode && node.Data == "p" {
					paragraphs++
				}
				for child := node.FirstChild; child != nil; child = child.NextSibling {
					visit(child)
				}
			}
			visit(root)
			if paragraphs != 1 {
				t.Fatalf("expected one paragraph, got %d", paragraphs)
			}
			org, err := document.Write(NewOrgWriter())
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(org, test.input) {
				t.Fatalf("Org writer changed the macro call or surrounding text: %s", org)
			}
		})
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
