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
	Title     string          `json:"title"`      // Human-readable title.
	Slug      string          `json:"slug"`       // URL-slug.
	UpdatedAt string          `json:"updated_at"` // Body update time.
	Image     string          `json:"image"`      // Path/URL/Placeholder to image.
	Body      string          `json:"body"`       // Main content of the page.
	Sidebar   string          `json:"sidebar"`    // The sidebar of the page.
	Anchors   []anchor.Anchor `json:"anchors"`    // The list of anchors to headers in the body.
	Children  *Node           `json:"children,omitempty"`
}

type Node struct {
	Slug     string           `json:"slug"`
	Title    string           `json:"title"`
	Children map[string]*Node `json:"children,omitempty"`
}

func NewNode(slug, title string) *Node {
	return &Node{Slug: slug, Title: title, Children: make(map[string]*Node)}
}

func (f *Node) getNode(slug string) *Node {
	if nextN, ok := f.Children[slug]; ok {
		return nextN
	} else {
		log.Fatalf("Expected nested Node %v in %v\n", slug, f.Slug)
	}
	return &Node{} // cannot happen
}

func (n *Node) hasNode(slug string) bool {
	_, ok := n.Children[slug]
	return ok
}

func (f *Node) AddNode(path []string, title string) {
	if len(path) == 0 {
		if f.Title == "" {
			f.Title = title
		}
		return
	}
	slug := path[0]
	log.Println(slug, path)
	if f.hasNode(slug) {
		f.getNode(slug).AddNode(path[1:], title)
		return
	}
	f.Children[slug] = NewNode(slug, title)
	f.getNode(slug).AddNode(path[1:], title)
}

func (f *Node) String() (str string) {
	str += f.Slug + ":" + f.Title + " + {\n"
	for _, child := range f.Children {
		str += "\t" + child.String()
	}
	str += "\n}"
	return str
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
			return nil, nil
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
		Image string
		Title string
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
		Body:      body,
		Sidebar:   sidebar,
		Anchors:   anchs,
	}, nil
}
