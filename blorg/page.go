package blorg

import (
	"fmt"
	"html/template"
	"os"
	"time"

	"github.com/niklasfasching/go-org/org"
)

type Page struct {
	*Config
	Document       *org.Document
	Info           os.FileInfo
	PermaLink      string
	Date           time.Time
	Content        template.HTML
	BufferSettings map[string]string
}

func NewPage(c *Config, path string, info os.FileInfo) (*Page, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	d := c.OrgConfig.Parse(f, path)
	content, err := d.Write(getWriter())
	if err != nil {
		return nil, err
	}
	date, err := time.Parse("2006-01-02", d.Get("DATE"))
	if err != nil {
		date, _ = time.Parse("2006-01-02", "1970-01-01")
	}
	return &Page{
		Config:         c,
		Document:       d,
		Info:           info,
		Date:           date,
		Content:        template.HTML(content),
		BufferSettings: d.BufferSettings,
	}, nil
}

func (p *Page) Render(path string) error {
	if p.BufferSettings["DRAFT"] != "" {
		return nil
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	templateName := "item"
	if v, ok := p.BufferSettings["TEMPLATE"]; ok {
		templateName = v
	}
	t := p.Template.Lookup(templateName)
	if t == nil {
		return fmt.Errorf("cannot render page %s: unknown template %s", p.Info.Name(), templateName)
	}
	return t.Execute(f, toMap(p.BufferSettings, p))
}

func (p *Page) Summary() template.HTML {
	for _, n := range p.Document.Nodes {
		switch n := n.(type) {
		case org.Block:
			if n.Name == "SUMMARY" {
				w := getWriter()
				org.WriteNodes(w, n.Children...)
				return template.HTML(w.String())
			}
		}
	}
	for i, n := range p.Document.Nodes {
		switch n.(type) {
		case org.Headline:
			w := getWriter()
			org.WriteNodes(w, p.Document.Nodes[:i]...)
			return template.HTML(w.String())
		}
	}
	return ""
}
