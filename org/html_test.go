package org

import (
	"strings"
	"testing"
)

func TestHTMLWriter(t *testing.T) {
	for _, path := range orgTestFiles() {
		expected := fileString(path[:len(path)-len(".org")] + ".html")
		reader, writer := strings.NewReader(fileString(path)), NewHTMLWriter()
		actual, err := NewDocument().SetPath(path).Parse(reader).Write(writer)
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
