package pages

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/datasektionen/taitan/anchor"

	"github.com/BurntSushi/toml"
	log "github.com/Sirupsen/logrus"
	"github.com/russross/blackfriday"
)

// Resp is the response we serve for file queries.
type Resp struct {
	Title     string          `json:"title"` // Human-readable title.
	Slug      string          `json:"slug"`  // URL-slug.
	URL       string          `json:"url"`
	UpdatedAt string          `json:"updated_at"` // Body update time.
	Image     string          `json:"image"`      // Path/URL/Placeholder to image.
	Message   string          `json:"message"`    // Message to show at top
	Body      string          `json:"body"`       // Main content of the page.
	Sidebar   string          `json:"sidebar"`    // The sidebar of the page.
	Anchors   []anchor.Anchor `json:"anchors"`    // The list of anchors to headers in the body.
	Nav       []*Node         `json:"nav,omitempty"`
}

// Node is a recursive node in a page tree.
type Node struct {
	path     string
	Slug     string  `json:"slug"`
	Title    string  `json:"title"`
	Active   bool    `json:"active,omitempty"`
	Expanded bool    `json:"expanded,omitempty"`
	Nav      []*Node `json:"nav,omitempty"`
}

// NewNode creates a new node with it's path, slug and page title.
func NewNode(path, slug, title string) *Node {
	return &Node{path: path, Slug: slug, Title: title, Nav: make([]*Node, 0)}
}

func (n *Node) getNode(path string) *Node {
	for _, c := range n.Nav {
		if c.path == path {
			return c
		}
	}
	log.Fatalf("Expected nested Node %v in %v\n", path, n.path)
	return &Node{} // cannot happen
}

func (n *Node) hasNode(path string) bool {
	for _, c := range n.Nav {
		if c.path == path {
			return true
		}
	}
	return false
}

// AddNode adds a node to the node tree.
func (n *Node) AddNode(root []string, p string, title string, paths []string, active bool, expanded bool) {
	// Yay! Create us!
	if len(paths) == 0 {
		n.Active = active
		n.Expanded = expanded
		n.Title = title
		n.Slug = p
		return
	}
	// Parent folder.
	parent := paths[0]

	// Have we already created the parent?
	if n.hasNode(parent) {
		if len(root) == 0 {
			return
		}
		if root[0] == parent {
			n.getNode(parent).AddNode(root[1:], p, title, paths[1:], false, false)
		} else if len(paths) == 1 {
			n.getNode(parent).AddNode(root, p, title, []string{}, false, false)
		}
		return
	}
	// Create it and move on.
	n.Nav = append(n.Nav, NewNode(parent, p, title))
	n.getNode(parent).AddNode(root, p, title, paths[1:], len(root) == 1 && root[0] == parent, len(root) > 1 && root[0] == parent)
}

// Num returns the recursive number of pages under this node.
func (n *Node) Num() int {
	sum := 1
	for _, c := range n.Nav {
		sum += c.Num()
	}
	return sum
}

// Load intializes a root directory and serves all sub-folders.
func Load(root string) (pages map[string]*Resp, err error) {
	var dirs []string
	err = filepath.Walk(root, func(path string, fi os.FileInfo, err error) error {
		// We only search for article directories.
		if !fi.IsDir() {
			return nil
		}

		// Ignore our .git folder.
		if fi.IsDir() && fi.Name() == ".git" {
			return filepath.SkipDir
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
	return filepath.Clean(strings.Replace(dir, root, "/", 1))
}

// parseDirs parses each directory into a response. Returns a map from requested
// urls into responses.
func parseDirs(root string, dirs []string) (pages map[string]*Resp, err error) {
	pages = map[string]*Resp{}
	for _, dir := range dirs {
		r, err := parseDir(root, dir)
		if err != nil {
			log.Warnln(err)
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
		Extensions: blackfriday.EXTENSION_AUTO_HEADER_IDS | blackfriday.EXTENSION_TABLES | blackfriday.EXTENSION_FENCED_CODE,
	})
	return string(buf), nil
}

// parseDir creates a response for a directory.
func parseDir(root, dir string) (*Resp, error) {
	log.WithField("dir", dir).Debug("Parsing directory:")

	// Our content files.
	var (
		bodyPath    = filepath.Join(dir, "body.md")
		sidebarPath = filepath.Join(dir, "sidebar.md")
		metaPath    = filepath.Join(dir, "meta.toml")
	)

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
	anchs, err := anchor.Anchors(body)
	if err != nil {
		return nil, err
	}

	// Parse meta data from a toml file.
	var meta struct {
		Image   string
		Title   string
		Message string
	}
	if _, err := toml.DecodeFile(metaPath, &meta); err != nil {
		return nil, err
	}

	const iso8601DateTime = "2006-01-02T15:04:05Z"
	return &Resp{
		Title:     meta.Title,
		Slug:      filepath.Base(stripRoot(root, dir)),
		UpdatedAt: fi.ModTime().Format(iso8601DateTime),
		Image:     meta.Image,
		Message:   meta.Message,
		Body:      body,
		Sidebar:   sidebar,
		Anchors:   anchs,
	}, nil
}
