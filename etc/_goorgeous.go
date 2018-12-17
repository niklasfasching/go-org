package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/chaseadamsio/goorgeous"
	"github.com/russross/blackfriday"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.Print(err)
		}
	}()
	log.SetFlags(0)
	log.SetOutput(os.Stdout)

	path := os.Args[1]
	in, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	flags := blackfriday.HTML_USE_XHTML
	flags |= blackfriday.LIST_ITEM_BEGINNING_OF_LIST
	flags |= blackfriday.HTML_FOOTNOTE_RETURN_LINKS
	parameters := blackfriday.HtmlRendererParameters{}
	parameters.FootnoteReturnLinkContents = "<sup>↩</sup>"
	renderer := blackfriday.HtmlRendererWithParameters(flags, "", "", parameters)
	log.Print(string(goorgeous.Org(in, renderer)))
}
