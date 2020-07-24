// blorg is a very minimal and broken static site generator. Don't use this. I initially wrote go-org to use Org mode in hugo
// and non crazy people should keep using hugo. I just like the idea of not having dependencies / following 80/20 rule. And blorg gives me what I need
// for a blog in a fraction of the LOC (hugo is a whooping 80k+ excluding dependencies - this will very likely stay <5k).
package blorg

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/niklasfasching/go-org/org"
)

type Config struct {
	ConfigFile string
	ContentDir string
	PublicDir  string
	Address    string
	BaseUrl    string
	Template   *template.Template
	OrgConfig  *org.Configuration
}

var DefaultConfigFile = "blorg.org"

var DefaultConfig = `
#+CONTENT: content
#+PUBLIC: public

* templates
** item
#+name: item
#+begin_src html
{{ . }}
#+end_src

** list
#+name: list
#+begin_src html
{{ . }}
#+end_src`

var TemplateFuncs = map[string]interface{}{
	"Slugify": slugify,
}

func ReadConfig(configFile string) (*Config, error) {
	baseUrl, address, publicDir, contentDir, workingDir := "/", ":3000", "public", "content", filepath.Dir(configFile)
	f, err := os.Open(configFile)
	if err != nil {
		return nil, err
	}
	orgConfig := org.New()
	document := orgConfig.Parse(f, configFile)
	if document.Error != nil {
		return nil, document.Error
	}
	m := document.BufferSettings
	if !strings.HasSuffix(m["BASE_URL"], "/") {
		m["BASE_URL"] += "/"
	}
	if v, exists := m["AUTO_LINK"]; exists {
		orgConfig.AutoLink = v == "true"
		delete(m, "AUTO_LINK")
	}
	if v, exists := m["ADDRESS"]; exists {
		address = v
		delete(m, "ADDRESS")
	}
	if _, exists := m["BASE_URL"]; exists {
		baseUrl = m["BASE_URL"]
	}
	if v, exists := m["PUBLIC"]; exists {
		publicDir = v
		delete(m, "PUBLIC")
	}
	if v, exists := m["CONTENT"]; exists {
		contentDir = v
		delete(m, "CONTENT")
	}
	if v, exists := m["MAX_EMPHASIS_NEW_LINES"]; exists {
		i, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("MAX_EMPHASIS_NEW_LINES: %v %w", v, err)
		}
		orgConfig.MaxEmphasisNewLines = i
		delete(m, "MAX_EMPHASIS_NEW_LINES")
	}

	for k, v := range m {
		if k == "OPTIONS" {
			orgConfig.DefaultSettings[k] = v + " " + orgConfig.DefaultSettings[k]
		} else {
			orgConfig.DefaultSettings[k] = v
		}
	}

	config := &Config{
		ConfigFile: configFile,
		ContentDir: filepath.Join(workingDir, contentDir),
		PublicDir:  filepath.Join(workingDir, publicDir),
		Address:    address,
		BaseUrl:    baseUrl,
		Template:   template.New("_").Funcs(TemplateFuncs),
		OrgConfig:  orgConfig,
	}
	for name, node := range document.NamedNodes {
		if block, ok := node.(org.Block); ok {
			if block.Parameters[0] != "html" {
				continue
			}
			if _, err := config.Template.New(name).Parse(org.String(block.Children)); err != nil {
				return nil, err
			}
		}
	}
	return config, nil
}

func (c *Config) Serve() error {
	http.Handle("/", http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if strings.HasSuffix(req.URL.Path, ".html") || strings.HasSuffix(req.URL.Path, "/") {
			start := time.Now()
			if c, err := ReadConfig(c.ConfigFile); err != nil {
				log.Fatal(err)
			} else {
				if err := c.Render(); err != nil {
					log.Fatal(err)
				}
			}
			log.Printf("render took %s", time.Since(start))
		}
		http.ServeFile(res, req, filepath.Join(c.PublicDir, path.Clean(req.URL.Path)))
	}))
	log.Printf("listening on: %s", c.Address)
	return http.ListenAndServe(c.Address, nil)
}

func (c *Config) Render() error {
	if err := os.RemoveAll(c.PublicDir); err != nil {
		return err
	}
	if err := os.MkdirAll(c.PublicDir, os.ModePerm); err != nil {
		return err
	}
	pages, err := c.RenderContent()
	if err != nil {
		return err
	}
	return c.RenderLists(pages)
}

func (c *Config) RenderContent() ([]*Page, error) {
	pages := []*Page{}
	err := filepath.Walk(c.ContentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(c.ContentDir, path)
		if err != nil {
			return err
		}
		publicPath := filepath.Join(c.PublicDir, relPath)
		publicInfo, err := os.Stat(publicPath)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		if info.IsDir() {
			return os.MkdirAll(publicPath, info.Mode())
		}
		if filepath.Ext(path) != ".org" && (os.IsNotExist(err) || info.ModTime().After(publicInfo.ModTime())) {
			return os.Link(path, publicPath)
		}
		p, err := NewPage(c, path, info)
		if err != nil {
			return err
		}
		pages = append(pages, p)

		p.PermaLink = c.BaseUrl + relPath[:len(relPath)-len(".org")] + ".html"
		return p.Render(publicPath[:len(publicPath)-len(".org")] + ".html")
	})
	sort.Slice(pages, func(i, j int) bool { return pages[i].Date.After(pages[j].Date) })
	return pages, err
}

func (c *Config) RenderLists(pages []*Page) error {
	ms := toMap(c.OrgConfig.DefaultSettings, nil)
	ms["Pages"] = pages
	lists := map[string]map[string][]interface{}{"": map[string][]interface{}{"": nil}}
	for _, p := range pages {
		if p.BufferSettings["DRAFT"] != "" {
			continue
		}
		mp := toMap(p.BufferSettings, p)
		if p.BufferSettings["DATE"] != "" {
			lists[""][""] = append(lists[""][""], mp)
		}
		for k, v := range p.BufferSettings {
			if strings.HasSuffix(k, "[]") {
				list := strings.ToLower(k[:len(k)-2])
				if lists[list] == nil {
					lists[list] = map[string][]interface{}{}
				}
				for _, sublist := range strings.Fields(v) {
					lists[list][sublist] = append(lists[list][strings.ToLower(sublist)], mp)
				}
			}
		}
	}
	for list, sublists := range lists {
		for sublist, pages := range sublists {
			ms["Title"] = strings.Title(sublist)
			ms["Pages"] = pages
			if err := c.RenderList(list, sublist, ms); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Config) RenderList(list, sublist string, m map[string]interface{}) error {
	t := c.Template.Lookup(list)
	if list == "" {
		m["Title"] = c.OrgConfig.DefaultSettings["TITLE"]
		t = c.Template.Lookup("index")
	}
	if t == nil {
		t = c.Template.Lookup("list")
	}
	if t == nil {
		return fmt.Errorf("cannot render list: neither template %s nor list", list)
	}
	path := filepath.Join(c.PublicDir, slugify(list), slugify(sublist))
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}
	f, err := os.Create(filepath.Join(path, "index.html"))
	if err != nil {
		return err
	}
	defer f.Close()
	return t.Execute(f, m)
}
