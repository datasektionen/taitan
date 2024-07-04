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
	Name  string `json:"name"`            // title
	Str   string `json:"str"`             // slug
	Color string `json:"color,omitempty"` // wtf
	Href  string `json:"href"`            // fmt.Sprintf("http://datasektionen.se%s", slug)
}

// NewFile returns a fuzzyfile.
func NewFile(resp map[string]*pages.Page) File {
	fs := make([]Fuzz, 0, 128)
	for path, r := range resp {
		title, ok := r.Titles[""]
		if !ok {
			for _, title_ := range r.Titles {
				title = title_
				break
			}
		}

		fs = append(fs, Fuzz{
			Name: title,
			Str:  r.Slug,
			Href: fmt.Sprintf("http://datasektionen.se%s", path),
		})
	}
	return File{
		Type:   "fuzzyfile",
		Fuzzes: fs,
	}
}
