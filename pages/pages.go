package pages

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/russross/blackfriday"
	"golang.org/x/net/html"
)

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

// Anchor is a html anchor tag with an id attribute and a value. Represents: <a
// id="Id">Value</a>
type Anchor struct {
	ID    string `json:"id"`    // Id of h2 element.
	Value string `json:"value"` // Value inside the anchor tag.
}

// Load intializes a root directory and serves all sub-folders.
func Load(root string) (pages map[string]*Resp, err error) {
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
		return nil, err
	}
	return parseDirs(root, dirs)
}

// stripRoot removes root level of a directory.
// This is because when a user requests:
// `/sektionen/om-oss` the actual path is: `root/sektionen/om-oss`
func stripRoot(root string, dir string) string {
	return strings.Replace(dir, root, "/", 1)
}

// parseDirs parses each directory into a response. Returns a map from requested
// urls into responses.
func parseDirs(root string, dirs []string) (pages map[string]*Resp, err error) {
	pages = map[string]*Resp{}
	for _, dir := range dirs {
		r, err := parseDir(dir)
		if err != nil {
			return nil, err
		}
		pages[stripRoot(root, dir)] = r
		log.WithFields(log.Fields{
			"Resp": r,
			"dir":  dir,
		}).Debug("Our parsed response\n")
	}
	return pages, nil
}

// toHTML reads a markdown file and returns a HTML string.
func toHTML(filename string) (string, error) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	// Use standard HTML rendering.
	renderer := blackfriday.HtmlRenderer(blackfriday.HTML_USE_XHTML, "", "")
	// Parse markdown where all id's are created from the values inside
	// the element tag.
	buf = blackfriday.MarkdownOptions(buf, renderer, blackfriday.Options{
		Extensions: blackfriday.EXTENSION_AUTO_HEADER_IDS,
	})
	return string(buf), nil
}

// parseDir creates a response for a directory.
func parseDir(dir string) (*Resp, error) {
	log.WithField("dir", dir).Debug("Current directory")

	// Our content files.
	bodyPath := filepath.Join(dir, "body.md")
	sidebarPath := filepath.Join(dir, "sidebar.md")

	// Parse markdown to HTML.
	body, err := toHTML(bodyPath)
	if err != nil {
		return nil, err
	}
	log.WithField("body", body).Debug("HTML of body.md")

	// Parse sidebar to HTML.
	sidebar, err := toHTML(sidebarPath)
	if err != nil {
		return nil, err
	}
	log.WithField("sidebar", sidebar).Debug("HTML of sidebar.md")

	// Parse modified at.
	fi, err := os.Stat(bodyPath)
	if err != nil {
		return nil, err
	}
	// Parse anchors in the body.
	anchs, err := anchors(body)
	if err != nil {
		return nil, err
	}
	// Title of the page (first anchor value).
	title := ""
	if len(anchs) > 0 {
		title = anchs[0].Value
	}
	const iso8601DateTime = "2006-01-02T15:04:05Z"
	return &Resp{
		Title:     title,
		Slug:      filepath.Base(dir),
		UpdatedAt: fi.ModTime().Format(iso8601DateTime),
		Image:     "unimplemented",
		Body:      body,
		Sidebar:   sidebar,
		Anchors:   anchs,
	}, nil
}

// anchors finds <h1> elements inside a HTML string to create a list of anchors.
func anchors(body string) (anchs []Anchor, err error) {
	node, err := html.Parse(strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	anchs = make([]Anchor, 0)
	// Recursively find <h1> elements.
	var findAnchors func(*html.Node)
	findAnchors = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "h1" {
			// Append valid anchors.
			anchs = anchor(n, anchs)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findAnchors(c)
		}
	}
	findAnchors(node)
	return anchs, nil
}

// anchor appends valid anchors to anchs.
func anchor(n *html.Node, anchs []Anchor) []Anchor {
	log.WithField("attrs", n.Attr).Debug("Found potential anchor (<h1>)")
	id := ""
	for _, attr := range n.Attr {
		if attr.Key == "id" {
			id = attr.Val
			break
		}
		return nil
	}
	val := ""
	if n.FirstChild != nil {
		val = plain(n)
	}
	return append(anchs, Anchor{
		ID:    id,
		Value: val,
	})
}

// Find plain text value of a HTML tag.
func plain(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		return plain(c)
	}
	return ""
}
