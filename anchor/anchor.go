package anchor

import (
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

// Anchor is a html anchor tag with an id attribute and a value. Represents: <a
// id="Id">Value</a>
// It is intended to link to titles in the document
type Anchor struct {
	ID          string `json:"id"`    // Id of the h* element.
	Value       string `json:"value"` // Value inside the anchor tag.
	HeaderLevel int    `json:"level"` // Level of the h* element.
}

// Anchors finds <h*> elements inside a HTML string to create a list of anchors.
func Anchors(body string) (anchs []Anchor, err error) {
	node, err := html.Parse(strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	anchs = make([]Anchor, 0)

	// Recursively find <h*> elements.
	var findAnchors func(*html.Node)
	findAnchors = func(n *html.Node) {
		if isHNode(n) {
			id := findIDAttr(n.Attr)
			val := plain(n)
			if id != "" && val != "" {
				headerLevel, _ := strconv.Atoi(n.Data[1:])
				anchs = append(anchs, Anchor{
					ID:          id,
					Value:       val,
					HeaderLevel: headerLevel,
				})
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findAnchors(c)
		}
	}
	findAnchors(node)
	return anchs, nil
}

func findIDAttr(attrs []html.Attribute) string {
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

func isHNode(n *html.Node) bool {
	if n.Type != html.ElementNode {
		return false
	}

	tags := []string{"h1", "h2", "h3", "h4", "h5", "h6"}
	for _, tag := range tags {
		if tag == n.Data {
			return true
		}
	}

	return false
}
