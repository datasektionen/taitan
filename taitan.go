package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/russross/blackfriday"
	"golang.org/x/net/html"
)

func init() {
	flag.Usage = usage
	flag.Parse()
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] ROOT\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}

// Pages maps slugs to responses.
var pages = map[string]*Resp{}

func main() {
	// We need a root folder.
	if flag.NArg() < 1 {
		usage()
	}

	// Recieve port from heruko.
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatalln("[!] $PORT environmental variable is not set.")
	}

	// Our root to read in markdown files.
	root := flag.Arg(0)
	log.Printf("[o] Root directory: `%s`\n", root)
	// We'll parse and store the responses ahead of time.
	err := loadRoot(root)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("[o] Pages: %#v\n", pages)

	log.Println("[o] Starting server.")
	log.Printf("[o] Listening on port: %s\n", port)

	// Our request handler.
	http.HandleFunc("/", handler)

	// Listen on port and serve with our handler.
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic(err)
	}
}

func loadRoot(root string) (err error) {
	var dirs []string
	err = filepath.Walk(root, func(path string, fi os.FileInfo, err error) error {
		// We only search for article directories.
		if !fi.IsDir() {
			return nil
		}
		dirs = append(dirs, path)
		return nil
	})
	if err != nil {
		return err
	}
	return parseDirs(dirs)
}

func stripRoot(dir string) string {
	i := strings.IndexRune(dir, '/')
	if i == -1 {
		return "/"
	}
	return dir[i:]
}

func parseDirs(dirs []string) (err error) {
	for _, dir := range dirs {
		r, err := parseDir(dir)
		if err != nil {
			return err
		}
		pages[stripRoot(dir)] = r
		fmt.Printf("[o] Resp: %#v\n", r)
	}
	return nil
}

func readMarkdown(filename string) (string, error) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	renderer := blackfriday.HtmlRenderer(blackfriday.HTML_USE_XHTML, "", "")
	return string(blackfriday.MarkdownOptions(buf, renderer, blackfriday.Options{
		Extensions: blackfriday.EXTENSION_AUTO_HEADER_IDS,
	})), nil
}

func parseDir(dir string) (*Resp, error) {
	log.Println("[o] dir:", dir)

	bodyPath := filepath.Join(dir, "body.md")
	log.Println("[o] Body path:", bodyPath)

	sidebarPath := filepath.Join(dir, "sidebar.md")
	log.Println("[o] Sidebar path:", sidebarPath)

	body, err := readMarkdown(bodyPath)
	if err != nil {
		return nil, err
	}
	log.Printf("[o] Body html: %#v\n", body)

	sidebar, err := readMarkdown(sidebarPath)
	if err != nil {
		return nil, err
	}
	log.Printf("[o] Sidebar html: %#v\n", sidebar)
	fi, err := os.Stat(bodyPath)
	if err != nil {
		return nil, err
	}
	anchs, err := anchors(body)
	if err != nil {
		return nil, err
	}
	title := ""
	if len(anchs) > 0 {
		title = anchs[0].Value
	}
	return &Resp{
		Title:     title,
		Slug:      filepath.Base(stripRoot(dir)),
		UpdatedAt: fi.ModTime().Format(ISO8601DateTime),
		Image:     "unimplemented",
		Body:      body,
		Sidebar:   sidebar,
		Anchors:   anchs,
	}, nil
}

const ISO8601DateTime = "2006-01-02T15:04:05Z"

func anchors(body string) (anchs []Anchor, err error) {
	node, err := html.Parse(strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	anchs = make([]Anchor, 0)
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "h1" {
			log.Println("[o] Add anchor:", n)
			id := ""
			for _, attr := range n.Attr {
				if attr.Key == "id" {
					id = attr.Val
					break
				}
				return
			}
			if n.FirstChild == nil {
				log.Println("[!] Empty value in node with id:", id)
				return
			}
			val := plain(n)
			anchs = append(anchs, Anchor{
				ID:    id,
				Value: val,
			})
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(node)
	return anchs, nil
}

func plain(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		return plain(c)
	}
	return ""
}

// Resp is the response we serve for file queries.
type Resp struct {
	Title     string   `json:"title"`      // Human-readable title.
	Slug      string   `json:"slug"`       // URL-slug.
	UpdatedAt string   `json:"updated_at"` // Body update time.
	Image     string   `json:"image"`      // Path/URL/Placeholder to image.
	Body      string   `json:"body"`
	Sidebar   string   `json:"sidebar"`
	Anchors   []Anchor `json:"anchors"`
}

// Anchor is a html anchor tag with an id attribute and a value. Represents: <a id="Id">Value</a>
type Anchor struct {
	ID    string `json:"id"`    // Id of h2 element.
	Value string `json:"value"` // Value inside the anchor tag.
}

// handler parses and serves responses to our file queries.
func handler(res http.ResponseWriter, req *http.Request) {
	// Requested URL. We extract the path.
	query := req.URL.Path
	log.Printf("[o] Got query: `%s`\n", query)

	clean := filepath.Clean(query)
	log.Printf("[o] Sanitized path: `%s`\n", clean)

	r, ok := pages[clean]
	if !ok {
		log.Printf("[!] Page doesn't exist: `%s`\n", clean)
		res.WriteHeader(404)
		return
	}
	log.Println("[o] Marshal the response.")
	buf, err := json.Marshal(r)
	if err != nil {
		log.Println("[!] Unexpected error:", err)
		res.WriteHeader(500)
		return
	}
	log.Println("[o] Serve the response.")
	fmt.Fprintln(res, string(buf))
}
