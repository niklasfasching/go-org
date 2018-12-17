package org

import (
	"regexp"
)

type Example struct {
	Children []Node
}

var exampleLineRegexp = regexp.MustCompile(`^(\s*): (.*)`)

func lexExample(line string) (token, bool) {
	if m := exampleLineRegexp.FindStringSubmatch(line); m != nil {
		return token{"example", len(m[1]), m[2], m}, true
	}
	return nilToken, false
}

func (d *Document) parseExample(i int, parentStop stopFn) (int, Node) {
	example, start := Example{}, i
	for ; !parentStop(d, i) && d.tokens[i].kind == "example"; i++ {
		example.Children = append(example.Children, Text{d.tokens[i].content, true})
	}
	return i - start, example
}
