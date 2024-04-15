package anchor

import (
	"bytes"
	"log"
	"slices"
	"testing"

	"golang.org/x/net/html"
)

var plaintests = []struct {
	in  *html.Node
	out string
}{
	{s2html("<b><b><b>plain</b></b></b>"), "plain"},
	{s2html("<b><b></b></b>"), ""},
	{s2html("asdf"), "asdf"},
}

func s2html(s string) *html.Node {
	n, err := html.Parse(bytes.NewBuffer([]byte(s)))
	if err != nil {
		log.Fatalln(err)
	}
	// Ignore DOM, <html>, <head> and <body> from html.Parse.
	return n.FirstChild.FirstChild.NextSibling.FirstChild
}

func TestPlain(t *testing.T) {
	for _, tt := range plaintests {
		got := plain(tt.in)
		if tt.out != got {
			t.Errorf("plain(%#v) => %q, want %q", tt.in, got, tt.out)
		}
	}
}

var anchortests = []struct {
	in  string
	out []Anchor
}{
	{`<h1 id="asdf"><b><b>plain</b></b></h1>`, []Anchor{{"asdf", "plain", 1}}},
	{`<b><b></b></b>`, []Anchor{}},
	{`asdf`, []Anchor{}},
	{`<h2 id="bing"><span style="color: red;">chilling</h2>`, []Anchor{{"bing", "chilling", 2}}},
}

func TestAnchors(t *testing.T) {
	for _, tt := range anchortests {
		got, err := Anchors(tt.in)
		if err != nil {
			t.Errorf("anchor(%v) returned error %q", tt.in, err)
		}
		if !slices.Equal(got, tt.out) {
			t.Errorf("anchor(%v) => %q, want %q", tt.in, got, tt.out)
		}
	}
}

var findattrtests = []struct {
	in  []html.Attribute
	out string
}{
	{[]html.Attribute{}, ""},
	{[]html.Attribute{{Key: "id", Val: "asdf"}}, "asdf"},
	{[]html.Attribute{{Key: "class", Val: "qwerty"}, {Key: "id", Val: "asdf"}}, "asdf"},
}

func TestFindIDAttr(t *testing.T) {
	for _, tt := range findattrtests {
		got := findIDAttr(tt.in)
		if got != tt.out {
			t.Errorf("findAttr(%v) => %q, want %q", tt.in, got, tt.out)
		}
	}
}
