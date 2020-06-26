package blorg

import (
	"reflect"
	"regexp"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/niklasfasching/go-org/org"
)

var snakeCaseRegexp = regexp.MustCompile(`(^[A-Za-z])|_([A-Za-z])`)
var whitespaceRegexp = regexp.MustCompile(`\s+`)
var nonWordCharRegexp = regexp.MustCompile(`[^\w-]`)

func toMap(bufferSettings map[string]string, x interface{}) map[string]interface{} {
	m := map[string]interface{}{}
	for k, v := range bufferSettings {
		k = toCamelCase(k)
		if strings.HasSuffix(k, "[]") {
			m[k[:len(k)-2]] = strings.Fields(v)
		} else {
			m[k] = v
		}
	}
	if x == nil {
		return m
	}
	v := reflect.ValueOf(x).Elem()
	for i := 0; i < v.NumField(); i++ {
		m[v.Type().Field(i).Name] = v.Field(i).Interface()
	}
	return m
}

func toCamelCase(s string) string {
	return snakeCaseRegexp.ReplaceAllStringFunc(strings.ToLower(s), func(s string) string {
		return strings.ToUpper(strings.Replace(s, "_", "", -1))
	})
}

func slugify(s string) string {
	s = strings.ToLower(s)
	s = whitespaceRegexp.ReplaceAllString(s, "-")
	s = nonWordCharRegexp.ReplaceAllString(s, "")
	return strings.Trim(s, "-")
}

func getWriter() org.Writer {
	w := org.NewHTMLWriter()
	w.HighlightCodeBlock = highlightCodeBlock
	return w
}

func highlightCodeBlock(source, lang string, inline bool) string {
	var w strings.Builder
	l := lexers.Get(lang)
	if l == nil {
		l = lexers.Fallback
	}
	l = chroma.Coalesce(l)
	it, _ := l.Tokenise(nil, source)
	_ = html.New().Format(&w, styles.Get("github"), it)
	if inline {
		return `<div class="highlight-inline">` + "\n" + w.String() + "\n" + `</div>`
	}
	return `<div class="highlight">` + "\n" + w.String() + "\n" + `</div>`
}
