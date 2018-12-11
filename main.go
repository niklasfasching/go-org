package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/niklasfasching/go-org/org"
)

func main() {
	log.SetFlags(0)
	log.SetOutput(os.Stdout)
	if len(os.Args) < 3 {
		log.Println("USAGE: org FILE OUTPUT_FORMAT")
		log.Fatal("supported output formats: org, html, html-chroma")
	}
	bs, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	r, out := bytes.NewReader(bs), ""
	switch strings.ToLower(os.Args[2]) {
	case "org":
		out = org.NewDocument().Parse(r).Write(org.NewOrgWriter()).String()
	case "html":
		out = org.NewDocument().Parse(r).Write(org.NewHTMLWriter()).String()
	case "html-chroma":
		writer := org.NewHTMLWriter()
		writer.HighlightCodeBlock = highlightCodeBlock
		out = org.NewDocument().Parse(r).Write(writer).String()
	default:
		log.Fatal("Unsupported output format")
	}
	log.Print(out)
}

func highlightCodeBlock(source, lang string) string {
	var w strings.Builder
	l := lexers.Get(lang)
	if l == nil {
		l = lexers.Fallback
	}
	l = chroma.Coalesce(l)
	it, _ := l.Tokenise(nil, source)
	_ = html.New().Format(&w, styles.Get("friendly"), it)
	return `<div class="highlight">` + w.String() + `</div>`
}
