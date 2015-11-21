package pages

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/datasektionen/taitan/anchor"

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
		r, err := parseDir(dir)
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
func parseDir(dir string) (*Resp, error) {
	log.WithField("dir", dir).Debug("Parsing directory:")

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
	anchs, err := anchor.Anchors(body)
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
