// Package importers provides functionality to track down packages depending on
// certain package.
package importers

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/webdevdata/webdevdata-tools/webdevdata"
	"golang.org/x/net/html"
)

// GoDoc is used to extract importers information from the godoc website.
// By default information is extracted from https://godoc.org
type GoDoc struct {
	// URL contains the url used to extract the information from.
	// If empty it'll use https://godoc.org
	URL string
}

// List looks for dependent packages tracked by godoc.org
func (g *GoDoc) List(pkg string, recursive bool) ([]string, error) {
	list := []string{}
	pkgs := []string{pkg}
	for len(pkgs) > 0 {
		pkg := pkgs[0]
		pkgs = pkgs[1:]
		l, err := g.fetchImportersList(pkg)
		if err != nil {
			return list, err
		}
		list = append(list, l...)
		if !recursive {
			continue
		}
		l, err = g.fetchSupackages(pkg)
		if err != nil {
			return list, err
		}
		pkgs = append(pkgs, l...)
	}
	return list, nil
}

func (g *GoDoc) url() (*url.URL, error) {
	u := g.URL
	if g.URL == "" {
		u = "https://godoc.org"
	}
	return url.Parse(u)
}

func (g *GoDoc) fetchImportersList(pkg string) ([]string, error) {
	u, err := g.url()
	if err != nil {
		return nil, err
	}
	u.Path = "/" + pkg
	u.RawQuery = "importers"
	return fetchList(u.String())
}

func (g *GoDoc) fetchSupackages(pkg string) ([]string, error) {
	u, err := g.url()
	if err != nil {
		return nil, err
	}
	u.Path = "/" + pkg
	return fetchList(u.String())
}

func fetchList(url string) ([]string, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("failed to fetch %s - %s", url, res.Status)
	}
	list := []string{}
	webdevdata.ProcessMatchingTagsReader(res.Body, "table a", func(node *html.Node) {
		p := strings.TrimLeft(webdevdata.GetAttr("href", node.Attr), "/")
		list = append(list, p)
	})
	return list, nil
}
