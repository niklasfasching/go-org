package org

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pmezard/go-difflib/difflib"
)

func TestOrgWriter(t *testing.T) {
	for _, path := range orgTestFiles() {
		expected := fileString(path)
		reader, writer := strings.NewReader(expected), NewOrgWriter()
		actual := NewDocument().Parse(reader).Write(writer).String()
		if actual != expected {
			t.Errorf("%s:\n%s'", path, diff(actual, expected))
		} else {
			t.Logf("%s: passed!", path)
		}
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

func fileString(path string) string {
	bs, err := ioutil.ReadFile(path)
	if err != nil {
		panic(fmt.Sprintf("Could not read file %s: %s", path, err))
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
