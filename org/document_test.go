package org

import (
	"reflect"
	"strings"
	"testing"
)

type frontMatterTest struct {
	name     string
	input    string
	handler  func(FrontMatter, string, string) error
	expected map[string]interface{}
}

var frontMatterTests = []frontMatterTest{
	{`basic`,
		`#+TITLE: The Title`,
		FrontMatterHandler,
		map[string]interface{}{"title": "The Title"}},
	{`empty`,
		`* No frontmatter here`,
		FrontMatterHandler,
		map[string]interface{}{}},
	{`custom handler`,
		`
        #+TITLE: The Title

        #+TAGS: foo bar

        `,
		func(fm FrontMatter, k, v string) error {
			switch k := strings.ToLower(k); k {
			case "title":
				fm[k] = "Thanks For All The Fish"
				return nil
			default:
				return FrontMatterHandler(fm, k, v)
			}
		},
		map[string]interface{}{
			"title": "Thanks For All The Fish",
			"tags":  []string{"foo", "bar"},
		}},
	{`multiple + ignored keyword`,
		`
         #+TITLE: The Title
         #+AUTHOR: The Author

         #+OTHER: some other keyword
         #+TAGS: this will become []string

         #+ALIASES: foo bar
         #+ALIASES: baz bam
         #+categories: foo bar

         something that's not a keyword or a text line without content

         #+SUBTITLE: The Subtitle`,
		FrontMatterHandler,
		map[string]interface{}{
			"title":      "The Title",
			"author":     "The Author",
			"other":      "some other keyword",
			"tags":       []string{"this", "will", "become", "[]string"},
			"aliases":    []string{"foo", "bar", "baz", "bam"},
			"categories": []string{"foo", "bar"},
		},
	},
}

func TestParseFrontMatter(t *testing.T) {
	for _, test := range frontMatterTests {
		actual, err := GetFrontMatter(strings.NewReader(test.input), test.handler)
		if err != nil {
			t.Errorf("%s\n got error: %s", test.name, err)
			continue
		}
		if !reflect.DeepEqual(test.expected, actual) {
			t.Errorf("%s\n got: %#v\nexpected: %#v\n%s'", test.name, actual, test.expected, test.input)
		}
	}
}
