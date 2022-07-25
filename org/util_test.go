package org

import (
	"fmt"
	"testing"
)

var parseRangesTests = map[string][][2]int{
	"3-5":           {{3, 5}},
	"3 8-10":        {{3, 3}, {8, 10}},
	"3   5 6":       {{3, 3}, {5, 5}, {6, 6}},
	" 9-10 5-6  3 ": {{9, 10}, {5, 6}, {3, 3}},
}

func TestParseRanges(t *testing.T) {
	for s, expected := range parseRangesTests {
		t.Run(s, func(t *testing.T) {
			actual := ParseRanges(s)
			// If this fails it looks like:
			// util_test.go:<line>:  9-10 5-6  3 :
			//     --- Actual
			//     +++ Expected
			//     @@ -1 +1 @@
			//     -[[9 10] [5 9] [3 3]]
			//     +[[9 10] [5 6] [3 3]]
			if len(actual) != len(expected) {
				t.Errorf("%v:\n%v", s, diff(fmt.Sprintf("%v", actual), fmt.Sprintf("%v", expected)))
			} else {
				for i := range actual {
					if actual[i] != expected[i] {
						t.Errorf("%v:\n%v", s, diff(fmt.Sprintf("%v", actual), fmt.Sprintf("%v", expected)))
					}
				}
			}
		})
	}

}
