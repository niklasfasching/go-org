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
	for _, path := range orgTestFiles() {
		expected := fileString(path[:len(path)-len(".org")] + ".html")
		reader, writer := strings.NewReader(fileString(path)), NewHTMLWriter()
		actual, err := New().Silent().Parse(reader, path).Write(writer)
		if err != nil {
			t.Errorf("%s\n got error: %s", path, err)
			continue
		}
		if actual != expected {
			t.Errorf("%s:\n%s'", path, diff(actual, expected))
		} else {
			t.Logf("%s: passed!", path)
		}
	}
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
