package org

import (
	"strconv"
	"strings"
)

func isSecondBlankLine(d *Document, i int) bool {
	if i-1 <= 0 {
		return false
	}
	t1, t2 := d.tokens[i-1], d.tokens[i]
	if t1.kind == "text" && t2.kind == "text" && t1.content == "" && t2.content == "" {
		return true
	}
	return false
}

func isImageOrVideoLink(n Node) bool {
	if l, ok := n.(RegularLink); ok && l.Kind() == "video" || l.Kind() == "image" {
		return true
	}
	return false
}

// Parse ranges like this:
// "3-5" -> [[3, 5]]
// "3 8-10" -> [[3, 3], [8, 10]]
// "3  5 6" -> [[3, 3], [5, 5], [6, 6]]
//
// This is Hugo's hlLinesToRanges with "startLine" removed and errors
// ignored.
func ParseRanges(s string) [][2]int {
	var ranges [][2]int
	s = strings.TrimSpace(s)
	if s == "" {
		return ranges
	}
	fields := strings.Split(s, " ")
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		numbers := strings.Split(field, "-")
		var r [2]int
		if len(numbers) > 1 {
			first, err := strconv.Atoi(numbers[0])
			if err != nil {
				return ranges
			}
			second, err := strconv.Atoi(numbers[1])
			if err != nil {
				return ranges
			}
			r[0] = first
			r[1] = second
		} else {
			first, err := strconv.Atoi(numbers[0])
			if err != nil {
				return ranges
			}
			r[0] = first
			r[1] = first
		}

		ranges = append(ranges, r)
	}
	return ranges
}

func IsNewLineChar(r rune) bool {
	return r == '\n' || r == '\r'
}
