// +build gofuzz

package org

import (
	"bytes"
	"strings"
)

// Fuzz function to be used by https://github.com/dvyukov/go-fuzz
func Fuzz(input []byte) int {
	d := NewDocument().Silent().Parse(bytes.NewReader(input))
	orgOutput, err := d.Write(NewOrgWriter())
	if err != nil {
		panic(err)
	}
	htmlOutputA, err := d.Write(NewHTMLWriter())
	if err != nil {
		panic(err)
	}
	htmlOutputB, err := NewDocument().Silent().Parse(strings.NewReader(orgOutput)).Write(NewHTMLWriter())
	if htmlOutputA != htmlOutputB {
		panic("rendered org results in different html than original input")
	}
	return 0
}
