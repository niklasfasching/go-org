package org

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pmezard/go-difflib/difflib"
)

type ExtendedOrgWriter struct {
	*OrgWriter
	callCount int
}

func (w *ExtendedOrgWriter) WriteText(t Text) {
	w.callCount++
	w.OrgWriter.WriteText(t)
}

func TestOrgWriter(t *testing.T) {
	testWriter(t, func() Writer { return NewOrgWriter() }, ".pretty_org")
}

func TestExtendedOrgWriter(t *testing.T) {
	p := Paragraph{Children: []Node{Text{Content: "text"}, Text{Content: "more text"}}}
	orgWriter := NewOrgWriter()
	extendedWriter := &ExtendedOrgWriter{orgWriter, 0}
	orgWriter.ExtendingWriter = extendedWriter
	WriteNodes(extendedWriter, p)
	if extendedWriter.callCount != 2 {
		t.Errorf("WriteText method of extending writer was not called: CallCount %d", extendedWriter.callCount)
	}
}

func testWriter(t *testing.T, newWriter func() Writer, ext string) {
	for _, path := range orgTestFiles() {
		tmpPath := path[:len(path)-len(".org")]
		t.Run(filepath.Base(tmpPath), func(t *testing.T) {
			expected := fileString(t, tmpPath+ext)
			reader := strings.NewReader(fileString(t, path))
			actual, err := New().Silent().Parse(reader, path).Write(newWriter())
			if err != nil {
				t.Fatalf("%s\n got error: %s", path, err)
			} else if actual != expected {
				t.Fatalf("%s:\n%s'", path, diff(actual, expected))
			}
		})
	}
}

func orgTestFiles() []string {
	dir := "./testdata"
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(fmt.Sprintf("Could not read directory: %s", err))
	}
	orgFiles := []string{}
	for _, f := range files {
		name := f.Name()
		if filepath.Ext(name) != ".org" {
			continue
		}
		orgFiles = append(orgFiles, filepath.Join(dir, name))
	}
	return orgFiles
}

func fileString(t *testing.T, path string) string {
	bs, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("Could not read file %s: %s", path, err)
	}
	return string(bs)
}

func diff(actual, expected string) string {
	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(actual),
		B:        difflib.SplitLines(expected),
		FromFile: "Actual",
		ToFile:   "Expected",
		Context:  3,
	}
	text, _ := difflib.GetUnifiedDiffString(diff)
	return text
}
