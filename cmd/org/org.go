package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/niklasfasching/org"
)

func main() {
	log.SetFlags(0)
	if len(os.Args) < 3 {
		log.Println("USAGE: org FILE OUTPUT_FORMAT")
		log.Fatal("supported output formats: org")
	}
	bs, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	r, out := bytes.NewReader(bs), ""
	switch strings.ToLower(os.Args[2]) {
	case "org":
		out = org.NewDocument().Parse(r).Write(org.NewOrgWriter()).String()
	default:
		log.Fatal("Unsupported output format")
	}
	log.Println(out)
}
