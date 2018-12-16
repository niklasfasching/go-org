package org

import (
	"regexp"
	"strings"
)

type Drawer struct {
	Name     string
	Children []Node
}

var beginDrawerRegexp = regexp.MustCompile(`^(\s*):(\S+):\s*$`)
var endDrawerRegexp = regexp.MustCompile(`^(\s*):END:\s*$`)

func lexDrawer(line string) (token, bool) {
	if m := endDrawerRegexp.FindStringSubmatch(line); m != nil {
		return token{"endDrawer", len(m[1]), "", m}, true
	} else if m := beginDrawerRegexp.FindStringSubmatch(line); m != nil {
		return token{"beginDrawer", len(m[1]), strings.ToUpper(m[2]), m}, true
	}
	return nilToken, false
}

func (d *Document) parseDrawer(i int, parentStop stopFn) (int, Node) {
	drawer, start := Drawer{Name: strings.ToUpper(d.tokens[i].content)}, i
	i++
	stop := func(d *Document, i int) bool {
		return parentStop(d, i) || d.tokens[i].kind == "endDrawer" || d.tokens[i].kind == "headline"
	}
	consumed, nodes := d.parseMany(i, stop)
	drawer.Children = nodes
	if d.tokens[i+consumed].kind == "endDrawer" {
		consumed++
	}
	return i + consumed - start, drawer
}
