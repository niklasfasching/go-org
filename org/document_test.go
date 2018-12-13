package org

import (
	"reflect"
	"strings"
	"testing"
)

type frontMatterTest struct {
	name     string
	input    string
	handler  func(string, string) interface{}
	expected map[string]interface{}
}

var frontMatterTests = []frontMatterTest{
	{`basic`,
		`#+TITLE: The Title`,
		DefaultFrontMatterHandler,
		map[string]interface{}{"TITLE": "The Title"}},
	{`empty`,
		`* No frontmatter here`,
		DefaultFrontMatterHandler,
		map[string]interface{}{}},
	{`custom handler`,
		`
        #+TITLE: The Title

        #+TAGS: foo bar

        `,
		func(k, v string) interface{} {
			switch k {
			case "TITLE":
				return "Thanks For All The Fish"
			default:
				return DefaultFrontMatterHandler(k, v)
			}
		},
		map[string]interface{}{
			"TITLE": "Thanks For All The Fish",
			"TAGS":  []string{"foo", "bar"},
		}},
	{`multiple + ignored keyword`,
		`
         #+TITLE: The Title
         #+AUTHOR: The Author

         #+OTHER: some other keyword
         #+TAGS: this will become []string

         something that's not a keyword or a text line without content

         #+SUBTITLE: The Subtitle`,
		DefaultFrontMatterHandler,
		map[string]interface{}{
			"TITLE":  "The Title",
			"AUTHOR": "The Author",
			"OTHER":  "some other keyword",
			"TAGS":   []string{"this", "will", "become", "[]string"},
		},
	},
}

func TestParseFrontMatter(t *testing.T) {
	for _, test := range frontMatterTests {
		actual := NewDocument().FrontMatter(strings.NewReader(test.input), test.handler)
		if !reflect.DeepEqual(test.expected, actual) {
			t.Errorf("%s\n got: %#v\nexpected: %#v\n%s'", test.name, actual, test.expected, test.input)
		}
	}
}
