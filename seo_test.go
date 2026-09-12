//go:build !postgres

package main

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/net/html"
	"sftrails/db"
)

func pageNodes(t *testing.T, body, tag string) []*html.Node {
	t.Helper()
	doc, err := html.Parse(strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	var nodes []*html.Node
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == tag {
			nodes = append(nodes, n)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return nodes
}

func attr(n *html.Node, name string) string {
	for _, a := range n.Attr {
		if a.Key == name {
			return a.Val
		}
	}
	return ""
}

func nodeText(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var b strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		b.WriteString(nodeText(c))
	}
	return b.String()
}

// Crawl the HTML links from the homepage and compare against the actual
// sitemap, verifying both discovery and each destination's search metadata.
func TestSearchPagesAreReachableAndCanonical(t *testing.T) {
	h := testRouter(t)
	get := func(path string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		return w
	}
	var sitemap struct {
		URLs []struct {
			Loc string `xml:"loc"`
		} `xml:"url"`
	}
	sm := get("/sitemap.xml")
	if err := xml.Unmarshal(sm.Body.Bytes(), &sitemap); err != nil {
		t.Fatal(err)
	}
	if len(sitemap.URLs) != 14 {
		t.Fatalf("unexpected sitemap size: %d", len(sitemap.URLs))
	}
	if strings.Contains(sm.Body.String(), "lastmod") {
		t.Fatal("sitemap invents modification dates")
	}
	home := get("/").Body.String()
	linked := map[string]bool{}
	for _, a := range pageNodes(t, home, "a") {
		linked[attr(a, "href")] = true
	}
	titles, descriptions := map[string]bool{}, map[string]bool{}
	for _, entry := range sitemap.URLs {
		path := strings.TrimPrefix(entry.Loc, "https://sftrails.info")
		if !linked[path] {
			t.Errorf("%s has no crawlable homepage link", path)
		}
		w := get(path)
		if w.Code != 200 || strings.Contains(w.Header().Get("X-Robots-Tag"), "noindex") {
			t.Fatalf("%s is not indexable: %d", path, w.Code)
		}
		body := w.Body.String()
		if got := len(pageNodes(t, body, "h1")); got != 1 {
			t.Errorf("%s: got %d main headings", path, got)
		}
		canonical := ""
		for _, link := range pageNodes(t, body, "link") {
			if attr(link, "rel") == "canonical" {
				canonical = attr(link, "href")
			}
		}
		if canonical != entry.Loc {
			t.Errorf("%s canonical = %s", path, canonical)
		}
		titleNodes := pageNodes(t, body, "title")
		if len(titleNodes) != 1 {
			t.Fatalf("%s: missing title", path)
		}
		title := nodeText(titleNodes[0])
		if title == "" || titles[title] {
			t.Errorf("%s: empty or duplicate title", path)
		}
		titles[title] = true
		desc := ""
		for _, meta := range pageNodes(t, body, "meta") {
			if attr(meta, "name") == "description" {
				desc = attr(meta, "content")
			}
		}
		if desc == "" || descriptions[desc] {
			t.Errorf("%s: empty or duplicate description", path)
		}
		descriptions[desc] = true
		for _, script := range pageNodes(t, body, "script") {
			if attr(script, "type") == "application/ld+json" && !json.Valid([]byte(nodeText(script))) {
				t.Errorf("%s: invalid structured data", path)
			}
		}
	}
}

func TestAlternateFormatsAndUtilityIndexing(t *testing.T) {
	h := testRouter(t)
	for _, path := range []string{"/", "/trail/markham-park", "/how-it-works"} {
		for _, format := range []string{"html", "accept", "suffix"} {
			url := path + "?source=seo"
			if format == "suffix" {
				url = path + ".md?source=seo"
				if path == "/" {
					url = "/index.md?source=seo"
				}
			}
			req := httptest.NewRequest("GET", url, nil)
			if format == "accept" {
				req.Header.Set("Accept", "text/markdown")
			}
			w := httptest.NewRecorder()
			h.ServeHTTP(w, req)
			if w.Code != 200 {
				t.Fatalf("%s: status %d", url, w.Code)
			}
			if !strings.Contains(w.Header().Get("Vary"), "Accept") {
				t.Errorf("%s: missing Vary Accept", url)
			}
			if format != "html" {
				links := strings.Join(w.Header().Values("Link"), ",")
				if !strings.Contains(links, `<https://sftrails.info`+path+`>; rel="canonical"`) {
					t.Errorf("%s: missing HTML canonical", url)
				}
				if !strings.Contains(w.Header().Get("Content-Type"), "text/markdown") {
					t.Errorf("%s: missing markdown", url)
				}
			}
		}
	}
	for _, path := range []string{"/trails-list", "/api/trails", "/api/trails/1"} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != 200 || w.Header().Get("X-Robots-Tag") != "noindex" {
			t.Errorf("%s: utility indexing exclusion failed", path)
		}
	}
	for _, path := range []string{"/trail/missing-trail", "/trail/missing-trail.md"} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != http.StatusNotFound {
			t.Errorf("%s: expected genuine 404", path)
		}
	}
}

func TestSitemapFailureDoesNotPublishPartialURLs(t *testing.T) {
	d, err := db.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Initialize(context.Background(), d); err != nil {
		t.Fatal(err)
	}
	h, err := newHTTPHandler(d, nil, "direct")
	if err != nil {
		t.Fatal(err)
	}
	d.Close()
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/sitemap.xml", nil))
	if w.Code != http.StatusServiceUnavailable || strings.Contains(w.Body.String(), "<urlset") {
		t.Fatal("database failure published a partial sitemap")
	}
}
