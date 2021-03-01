package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/niklasfasching/go-org/blorg"
	"github.com/niklasfasching/go-org/org"
)

var usage = `Usage: go-org COMMAND [ARGS]...
Commands:
- render FILE FORMAT
  FORMAT: org, html, html-chroma
- blorg
  - blorg init
  - blorg build
  - blorg serve
`

func main() {
	log.SetFlags(0)
	if len(os.Args) < 2 {
		log.Fatal(usage)
	}
	switch cmd, args := os.Args[1], os.Args[2:]; cmd {
	case "render":
		render(args)
	case "blorg":
		runBlorg(args)
	default:
		log.Fatal(usage)
	}
}

func runBlorg(args []string) {
	if len(args) == 0 {
		log.Fatal(usage)
	}
	switch strings.ToLower(args[0]) {
	case "init":
		if _, err := os.Stat(blorg.DefaultConfigFile); !os.IsNotExist(err) {
			log.Fatalf("%s already exists", blorg.DefaultConfigFile)
		}
		if err := ioutil.WriteFile(blorg.DefaultConfigFile, []byte(blorg.DefaultConfig), os.ModePerm); err != nil {
			log.Fatal(err)
		}
		if err := os.MkdirAll("content", os.ModePerm); err != nil {
			log.Fatal(err)
		}
		log.Println("./blorg.org and ./content/ created. Please adapt ./blorg.org")
	case "build":
		config, err := blorg.ReadConfig(blorg.DefaultConfigFile)
		if err != nil {
			log.Fatal(err)
		}
		if err := config.Render(); err != nil {
			log.Fatal(err)
		}
		log.Println("blorg build finished")
	case "serve":
		config, err := blorg.ReadConfig(blorg.DefaultConfigFile)
		if err != nil {
			log.Fatal(err)
		}
		log.Fatal(config.Serve())
	default:
		log.Fatal(usage)
	}
}

func render(args []string) {
	if len(args) < 2 {
		log.Fatal(usage)
	}
	path, format := args[0], args[1]
	bs, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	d := org.New().Parse(bytes.NewReader(bs), path)
	write := func(w org.Writer) {
		out, err := d.Write(w)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Fprint(os.Stdout, out)
	}
	switch strings.ToLower(format) {
	case "org":
		write(org.NewOrgWriter())
	case "html":
		write(org.NewHTMLWriter())
	case "html-chroma":
		writer := org.NewHTMLWriter()
		writer.HighlightCodeBlock = highlightCodeBlock
		write(writer)
	default:
		log.Fatal(usage)
	}
}

func highlightCodeBlock(source, lang string, inline bool) string {
	var w strings.Builder
	l := lexers.Get(lang)
	if l == nil {
		l = lexers.Fallback
	}
	l = chroma.Coalesce(l)
	it, _ := l.Tokenise(nil, source)
	_ = html.New().Format(&w, styles.Get("friendly"), it)
	if inline {
		return `<div class="highlight-inline">` + "\n" + w.String() + "\n" + `</div>`
	}
	return `<div class="highlight">` + "\n" + w.String() + "\n" + `</div>`
}
