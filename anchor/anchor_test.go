package anchor

import (
	"bytes"
	"fmt"
	"log"
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
	in  *html.Node
	in2 []Anchor
	out []Anchor
}{
	{s2html(`<h1 id="asdf"><b><b>plain</b></b></h1>`), []Anchor{}, []Anchor{{"asdf", "plain", 1}}},
	{s2html(`<b><b></b></b>`), []Anchor{{"qwerty", "asdf", 1}}, []Anchor{{"qwerty", "asdf", 1}}},
	{s2html(`asdf`), []Anchor{}, []Anchor{{"", "asdf", 1}}},
}

func TestAnchor(t *testing.T) {
	for _, tt := range anchortests {
		got := anchor(tt.in, tt.in2)
		if fmt.Sprintf("%#v\n", got) != fmt.Sprintf("%#v\n", tt.out) {
			t.Errorf("anchor(%v, []Anchor{}) => %q, want %q", tt.in, got, tt.out)
		}
	}
}

var findattrtests = []struct {
	in  string
	in2 []html.Attribute
	out string
}{
	{"id", []html.Attribute{}, ""},
	{"id", []html.Attribute{{Key: "id", Val: "asdf"}}, "asdf"},
	{"id", []html.Attribute{{Key: "class", Val: "qwerty"}, {Key: "id", Val: "asdf"}}, "asdf"},
}

func TestFindAttr(t *testing.T) {
	for _, tt := range findattrtests {
		got := findAttr(tt.in, tt.in2)
		if got != tt.out {
			t.Errorf("findAttr(%v, %v) => %q, want %q", tt.in, tt.in2, got, tt.out)
		}
	}
}
