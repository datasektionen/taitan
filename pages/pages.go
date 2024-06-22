package pages

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/datasektionen/taitan/anchor"

	"github.com/BurntSushi/toml"
	"github.com/russross/blackfriday"
	log "github.com/sirupsen/logrus"
)

type LangLookup map[string]string
type LangAnchorLookup map[string][]anchor.Anchor

type Page struct {
	Titles    LangLookup       // Human-readable title.
	Slug      string           // URL-slug.
	URL       string           // Actual url?
	UpdatedAt LangLookup       // Page update time.
	Image     string           // Path/URL/Placeholder to image.
	Message   string           // Message to show at top
	Bodies    LangLookup       // Main content of the page.
	Sidebars  LangLookup       // The sidebar of the page.
	Sort      *int             // The order that the tab should appear in on the page
	Expanded  bool             // Should the Nav-tree rooted in this node always be expanded one step when loaded?
	Anchors   LangAnchorLookup // The list of anchors to headers in the body.
}

// Node is a recursive node in a page tree.
type Node struct {
	path     string
	Slug     string  `json:"slug"`
	Title    string  `json:"title"`
	Active   bool    `json:"active,omitempty"`
	Expanded bool    `json:"expanded,omitempty"`
	Sort     *int    `json:"sort,omitempty"`
	Nav      []*Node `json:"nav,omitempty"`
}

// Meta defines the attributes to be loaded from the meta.toml file
type Meta struct {
	Image     string
	Title     string
	Message   string
	Sort      *int
	Expanded  bool
	Sensitive bool
}

const (
	metaFile        = "meta.toml"
	iso8601DateTime = "2006-01-02T15:04:05Z"
)

var (
	bodyReg    = regexp.MustCompile("body(_\\w+)\\.md")
	sidebarReg = regexp.MustCompile("sidebar(_\\w+)\\.md")
	titleReg   = regexp.MustCompile("Title(_\\w+)")
)

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
func (n *Node) AddNode(root []string, p string, title string, paths []string, active bool, expanded bool, sort *int) {
	// Yay! Create us!
	if len(paths) == 0 {
		n.Active = active
		n.Expanded = expanded
		n.Title = title
		n.Slug = p
		n.Sort = sort
		return
	}

	// Parent folder.
	parent := paths[0]

	// Have we already created the parent?
	if n.hasNode(parent) {
		if len(root) == 0 {
			if n.getNode(parent).Expanded {
				n.getNode(parent).AddNode(root, p, title, paths[1:], false, expanded, sort)
			}
			return
		}
		if root[0] == parent || n.getNode(parent).Expanded {
			n.getNode(parent).AddNode(root[1:], p, title, paths[1:], false, expanded, sort)
		} else if len(paths) == 1 {
			n.getNode(parent).AddNode(root, p, title, []string{}, false, expanded, sort)
		}
		return
	}
	// Create it and move on.
	n.Nav = append(n.Nav, NewNode(parent, p, title))
	n.getNode(parent).AddNode(
		root,
		p,
		title,
		paths[1:],
		len(root) == 1 && root[0] == parent,
		expanded || (len(root) > 1 && root[0] == parent),
		sort,
	)
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
func Load(isReception bool, root string) (map[string]*Page, error) {
	var dirs []string
	err := filepath.Walk(root, func(path string, fi os.FileInfo, err error) error {
		// We only search for article directories.
		if !fi.IsDir() {
			return nil
		}

		// Ignore our .git folder.
		if fi.IsDir() && fi.Name()[0] == '.' {
			return filepath.SkipDir
		}
		dirs = append(dirs, path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return parseDirs(isReception, root, dirs)
}

// stripRoot removes root level of a directory.
// This is because when a user requests:
// `/sektionen/om-oss` the actual path is: `root/sektionen/om-oss`
func stripRoot(root string, dir string) string {
	return filepath.Clean(strings.Replace(dir, root, "/", 1))
}

// parseDirs parses each directory into a response. Returns a map from requested
// urls into responses.
func parseDirs(isReception bool, root string, dirs []string) (map[string]*Page, error) {
	pages := make(map[string]*Page)
	for _, dir := range dirs {
		r, err := parseDir(isReception, root, dir)
		if err != nil {
			log.Warnln(err)
			return nil, err
		}
		if r == nil {
			continue
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
func toHTML(isReception bool, filename string) (string, error) {
	rawMarkdown, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	t, err := template.New("").Parse(string(rawMarkdown))
	if err != nil {
		return "", err
	}
	var filteredMarkdown bytes.Buffer
	if err := t.Execute(&filteredMarkdown, map[string]any{"reception": isReception}); err != nil {
		return "", err
	}
	// Use standard HTML rendering.
	renderer := blackfriday.HtmlRenderer(blackfriday.HTML_USE_XHTML, "", "")
	// Parse markdown where all id's are created from the values inside
	// the element tag.
	html := blackfriday.MarkdownOptions(filteredMarkdown.Bytes(), renderer, blackfriday.Options{
		Extensions: blackfriday.EXTENSION_AUTO_HEADER_IDS | blackfriday.EXTENSION_TABLES | blackfriday.EXTENSION_FENCED_CODE,
	})
	return string(html), nil
}

// parseDir creates a response for a directory.
func parseDir(isReception bool, root, dir string) (*Page, error) {
	log.WithField("dir", dir).Debug("Parsing directory:")

	bodies := make(LangLookup)
	sidebars := make(LangLookup)
	commitTimes := make(LangLookup)
	titles := make(LangLookup)
	anchorsLists := make(LangAnchorLookup)

	entries, err := os.ReadDir(dir)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		entryPath := filepath.Join(dir, entry.Name())

		if match := bodyReg.FindSubmatch([]byte(entry.Name())); match != nil {
			lang := ""
			if len(match) > 1 && len(match[1]) > 0 {
				lang = string(match[1][1:])
			}
			bodies[lang], err = toHTML(isReception, entryPath)
			log.WithField("body", bodies[lang]).Debug("HTML of body_" + lang + ".md")

			if err != nil {
				return nil, err
			}

			commitTime, err := getCommitTime(root, entryPath)
			if err != nil {
				commitTimes[lang] = time.Now().Format(iso8601DateTime)
			} else {
				commitTimes[lang] = commitTime.Format(iso8601DateTime)
			}

			// Parse anchors in the body.
			anchorsLists[lang], err = anchor.Anchors(bodies[lang])
			if err != nil {
				return nil, err
			}
		}

		if match := sidebarReg.FindSubmatch([]byte(entry.Name())); match != nil {
			lang := ""
			if len(match) > 1 && len(match[1]) > 0 {
				lang = string(match[1][1:])
			}
			sidebars[lang], err = toHTML(isReception, entryPath)
			log.WithField("sidebar", sidebars[lang]).Debug("HTML of sidebar" + lang + ".md")
			if err != nil {
				return nil, err
			}
		}
	}

	// Parse meta data from a toml file.
	metaPath := filepath.Join(dir, metaFile)
	var meta = Meta{
		Sort:     nil, // all pages without a sort-tag should be after the pages with a sort-tag, but should keep their internal order
		Expanded: false,
	}
	var metaMap = make(map[string]any)
	if _, err := toml.DecodeFile(metaPath, &meta); err != nil {
		return nil, err
	}
	if _, err := toml.DecodeFile(metaPath, &metaMap); err != nil {
		return nil, err
	}

	if meta.Sensitive && isReception {
		return nil, nil
	}
	for k, v := range metaMap {
		if match := titleReg.FindSubmatch([]byte(k)); match != nil {
			switch v.(type) {
			case string:
				lang := ""
				if len(match) > 1 && len(match[1]) > 0 {
					lang = string(match[1][1:])
				}
				titles[lang] = v.(string)
			}
		}
	}

		Titles:    titles,
	return &Page{
		Slug:      filepath.Base(stripRoot(root, dir)),
		UpdatedAt: commitTimes,
		Image:     meta.Image,
		Message:   meta.Message,
		Bodies:    bodies,
		Sidebars:  sidebars,
		Anchors:   anchorsLists,
		Expanded:  meta.Expanded,
		Sort:      meta.Sort,
	}, nil
}

// getCommitTime returns last commit time for a file.
func getCommitTime(root string, filePath string) (time.Time, error) {
	gitDir := fmt.Sprintf("--git-dir=%s/.git", root)
	// root/page/body.md => page/body.md
	filePath = filepath.Clean(strings.TrimPrefix(filePath, root+"/"))

	cmd := exec.Command("git", gitDir, "log", "-n1", "--format=%at", "--", filePath)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Printf("Git failed. Stderr: %s", strings.TrimSpace(stderr.String()))
		return time.Time{}, err
	}

	lastCommitTimestamp, err := strconv.ParseInt(strings.TrimSpace(stdout.String()), 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(lastCommitTimestamp, 0), nil
}
