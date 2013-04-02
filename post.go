package main

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

var (
	ErrBadPostName = errors.New("Invalid post name. Expecting format YYYY-MM-DD-name-of-post.markdown")
)

// ParseParse will parse a file with front-end YAML and markup content, and
// return a key-value Post structure.
func ParsePost(fn string) (Page, error) {
	post, err := ParsePage(fn)
	if err != nil {
		return nil, err
	}

	// parse the Date and Title from the post's file name
	_,f := filepath.Split(fn)
	t, d, err := parsePostName(f)
	if err != nil {
		return nil, err
	}

	// set the post's date and title
	// ignore the title if the user specified in the front-end yaml
	post["date"] = d
	if post.GetTitle() == "" {
		post["title"] = t
	}

	// figoure out the Posts permalink
	mon := fmt.Sprintf("%02d", d.Month())
	day := fmt.Sprintf("%02d", d.Day())
	year := fmt.Sprintf("%02d", d.Year())
	name := replaceExt(f, ".html")
	post["id"] = filepath.Join(year, mon, day, f) // TODO try to remember why I need this field
	post["url"]= filepath.Join(year, mon, day, name[11:])

	return post, nil
}

// Helper function to parse a blog posts filename, which is in the following
// format: YYYY-MM-DD-name-of-post.markdown
//
// the name of the post will be separated from the time of the post, both of
// which are returned by this function
func parsePostName(fn string) (name string, date time.Time, err error) {
	if len(fn) < 12 {
		err = ErrBadPostName
		return
	}
	date, err = time.Parse("2006-01-02", fn[:10])
	if err != nil {
		return
	}
	name = fn[11:]
	name = removeExt(name)

	name = strings.Replace(name, "-", " ", -1)
	name = strings.ToTitle(name)
	return
}
