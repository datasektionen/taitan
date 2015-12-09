package anchor

import (
	"strings"

	log "github.com/Sirupsen/logrus"
	"golang.org/x/net/html"
)

// Anchor is a html anchor tag with an id attribute and a value. Represents: <a
// id="Id">Value</a>
type Anchor struct {
	ID    string `json:"id"`    // Id of h2 element.
	Value string `json:"value"` // Value inside the anchor tag.
}

// Anchors finds <h1> elements inside a HTML string to create a list of anchors.
func Anchors(body string) (anchs []Anchor, err error) {
	node, err := html.Parse(strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	anchs = make([]Anchor, 0)
	// Recursively find <h1> elements.
	var findAnchors func(*html.Node)
	findAnchors = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "h2" {
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
	log.WithField("attrs", n.Attr).Debug("Found potential anchor (<h2>)")
	id := findAttr("id", n.Attr)
	val := plain(n)
	if val == "" && id == "" {
		return anchs
	}
	return append(anchs, Anchor{
		ID:    id,
		Value: val,
	})
}

func findAttr(key string, attrs []html.Attribute) string {
	for _, attr := range attrs {
		if attr.Key == "id" {
			return attr.Val
		}
	}
	return ""
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
