package fuzz

import (
	"fmt"

	"github.com/datasektionen/taitan/pages"
)

// File is a fuzzy file.
type File struct {
	Type   string `json:"@type"`
	Fuzzes []Fuzz `json:"fuzzes"`
}

// Fuzz is a files fuzzy metadata.
type Fuzz struct {
	Name  string `json:"name"`  // title
	Str   string `json:"str"`   // slug
	Color string `json:"color"` // wtf
	Href  string `json:"href"`  // fmt.Sprintf("http://datasektionen.se%s", slug)
}

// NewFile returns a fuzzyfile.
func NewFile(resp map[string]*pages.Resp) File {
	fs := make([]Fuzz, 0, 128)
	for path, r := range resp {
		fs = append(fs, Fuzz{
			Name:  r.Title,
			Str:   r.Slug,
			Color: "not implemented",
			Href:  fmt.Sprintf("http://datasektionen.se%s", path),
		})
	}
	return File{
		Type:   "fuzzyfile",
		Fuzzes: fs,
	}
}
