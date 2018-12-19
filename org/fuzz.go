// +build gofuzz

package org

import "bytes"

// Fuzz function to be used by https://github.com/dvyukov/go-fuzz
func Fuzz(data []byte) int {
	d := NewDocument().Silent().Parse(bytes.NewReader(data))
	_, err := d.Write(NewOrgWriter())
	if err != nil {
		panic(err)
	}
	_, err = d.Write(NewHTMLWriter())
	if err != nil {
		panic(err)
	}
	return 0
}
