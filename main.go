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

func main() {
	log := log.New(os.Stderr, "", 0)
	if len(os.Args) < 2 {
		log.Println("USAGE: org COMMAND [ARGS]")
		log.Println("- org render FILE OUTPUT_FORMAT")
		log.Println("  OUTPUT_FORMAT: org, html, html-chroma")
		log.Println("- org blorg init")
		log.Println("- org blorg build")
		log.Println("- org blorg serve")
		os.Exit(1)
	}

	switch cmd := strings.ToLower(os.Args[1]); cmd {
	case "render":
		if len(os.Args) < 4 {
			log.Fatal("USAGE: org render FILE OUTPUT_FORMAT")
		}
		out, err := render(os.Args[2], os.Args[3])
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
		fmt.Fprint(os.Stdout, out)
	case "blorg":
		if err := runBlorg(os.Args[2]); err != nil {
			log.Fatalf("Error: %v", err)
		}
	default:
		log.Fatalf("Unsupported command: %s", cmd)
	}
}

func runBlorg(cmd string) error {
	switch strings.ToLower(cmd) {
	case "init":
		if _, err := os.Stat(blorg.DefaultConfigFile); !os.IsNotExist(err) {
			return err
		}
		if err := ioutil.WriteFile(blorg.DefaultConfigFile, []byte(blorg.DefaultConfig), os.ModePerm); err != nil {
			return err
		}
		log.Printf("blorg init finished: Wrote ./%s", blorg.DefaultConfigFile)
		return nil
	case "build":
		config, err := blorg.ReadConfig(blorg.DefaultConfigFile)
		if err != nil {
			return err
		}
		if err := config.Render(); err != nil {
			return err
		}
		log.Println("blorg build finished")
		return nil
	case "serve":
		config, err := blorg.ReadConfig(blorg.DefaultConfigFile)
		if err != nil {
			return err
		}
		return config.Serve()
	default:
		return fmt.Errorf("Supported commands: init build serve")
	}
}

func render(path, format string) (string, error) {
	bs, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	d := org.New().Parse(bytes.NewReader(bs), path)
	switch strings.ToLower(format) {
	case "org":
		return d.Write(org.NewOrgWriter())
	case "html":
		return d.Write(org.NewHTMLWriter())
	case "html-chroma":
		writer := org.NewHTMLWriter()
		writer.HighlightCodeBlock = highlightCodeBlock
		return d.Write(writer)
	default:
		return "", fmt.Errorf("unsupported output format: %s", format)
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
