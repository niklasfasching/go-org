package org

import (
	"math"
	"regexp"
	"strings"
)

type Paragraph struct{ Children []Node }
type HorizontalRule struct{}

var horizontalRuleRegexp = regexp.MustCompile(`^(\s*)-{5,}\s*$`)
var plainTextRegexp = regexp.MustCompile(`^(\s*)(.*)`)

func lexText(line string) (token, bool) {
	if m := plainTextRegexp.FindStringSubmatch(line); m != nil {
		return token{"text", len(m[1]), m[2], m}, true
	}
	return nilToken, false
}

func lexHorizontalRule(line string) (token, bool) {
	if m := horizontalRuleRegexp.FindStringSubmatch(line); m != nil {
		return token{"horizontalRule", len(m[1]), "", m}, true
	}
	return nilToken, false
}

func (d *Document) parseParagraph(i int, parentStop stopFn) (int, RangedNode) {
	lines, start := []string{d.tokens[i].content}, i
	stop := func(d *Document, i int) bool {
		return parentStop(d, i) || d.tokens[i].kind != "text" || d.tokens[i].content == ""
	}
	for i += 1; !stop(d, i); i++ {
		lvl := math.Max(float64(d.tokens[i].lvl-d.baseLvl), 0)
		lines = append(lines, strings.Repeat(" ", int(lvl))+d.tokens[i].content)
	}
	iBlocks := d.fromRangedNodesToNodes(d.parseInline(strings.Join(lines, "\n")))
	return i - start, RangedNode{Paragraph{iBlocks}, start, i}
}

func (d *Document) parseHorizontalRule(i int, parentStop stopFn) (int, RangedNode) {
	return 1, RangedNode{HorizontalRule{}, i, i + 1}
}

func (n Paragraph) String() string      { return orgWriter.WriteNodesAsString(n) }
func (n HorizontalRule) String() string { return orgWriter.WriteNodesAsString(n) }
