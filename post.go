package main

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"
	"github.com/opesun/slugify"
)

var (
	ErrBadPostName = errors.New("Invalid post name. Expecting format YYYY-MM-DD-name-of-post.markdown")
)

func getPostUrl(title string, date time.Time, categories []string, permalink string) (url string) {
	switch permalink {
		case "date":
			permalink = "/:categories/:year/:month/:day/:title.html"
		case "pretty":
			permalink = "/:categories/:year/:month/:day/:title/"
		case "none":
			permalink = "/:categories/:title.html"
	}

	url = permalink
	url = strings.Replace(url, ":title", title, -1)
	url = strings.Replace(url, ":year", fmt.Sprintf("%02d", date.Year()), -1)
	url = strings.Replace(url, ":month", fmt.Sprintf("%02d", date.Month()), -1)
	url = strings.Replace(url, ":i_month", fmt.Sprintf("%d", date.Month()), -1)
	url = strings.Replace(url, ":day", fmt.Sprintf("%02d", date.Day()), -1)
	url = strings.Replace(url, ":i_day", fmt.Sprintf("%d", date.Day()), -1)

	for i, value := range categories {
		categories[i] = slugify.S(value)
	}
	url = strings.Replace(url, ":categories", strings.Join(categories, "/"), -1)
	url = strings.Replace(url, "//", "/", -1)
	return url
}

// ParseParse will parse a file with front-end YAML and markup content, and
// return a key-value Post structure.
func ParsePost(fn string, permalink string) (Page, error) {
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

	// Figure out the Posts permalink
	title := replaceExt(f, "")[11:]
	post["url"] = getPostUrl(title, d, post.GetCategories(), permalink)

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
