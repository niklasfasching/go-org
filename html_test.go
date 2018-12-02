package org

import (
	"strings"
	"testing"
)

func TestHTMLWriter(t *testing.T) {
	for _, path := range orgTestFiles() {
		reader, writer := strings.NewReader(fileString(path)), NewHTMLWriter()
		actual := NewDocument().Parse(reader).Write(writer).String()
		expected := fileString(path[:len(path)-len(".org")] + ".html")

		if expected != actual {
			t.Errorf("%s:\n%s'", path, diff(actual, expected))
		} else {
			t.Logf("%s: passed!", path)
		}
	}
}
