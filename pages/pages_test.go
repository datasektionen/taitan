package pages

import (
	"log"
	"testing"
)

var striptests = []struct {
	in, in2, out string
}{
	{"test/", "test/sektionen/", "/sektionen"},
	{"test", "test", "/"},
	{"../test/", "../test/", "/"},
	{"../test/", "../test/sektionen", "/sektionen"},
	{"/", "/", "/"},
}

func TestStripRoot(t *testing.T) {
	for _, tt := range striptests {
		got := stripRoot(tt.in, tt.in2)
		if tt.out != got {
			t.Errorf("stripRoot(%q, %q) => %q, want %q", tt.in, tt.in2, got, tt.out)
		}
	}
}

var toHTMLtests = []struct {
	in, out string
}{
	{"test/body.md", "<h1 id=\"id-test\">Id test</h1>\n"},
}

func TestToHTML(t *testing.T) {
	for _, tt := range toHTMLtests {
		got, err := toHTML(false, tt.in)
		if err != nil {
			log.Fatalln(err)
		}
		if tt.out != got {
			t.Errorf("toHTML(%q) => %q, want %q", tt.in, got, tt.out)
		}
	}
}
