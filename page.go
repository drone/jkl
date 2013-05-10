package main

import (
	"bytes"
	"fmt"
	"github.com/russross/blackfriday"
	"github.com/wendal/goyaml2"
	"io"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
)

// A Page represents the key-value pairs in a page or posts front-end YAML as
// well as the markup in the body.
type Page map[string]interface{}

// ParsePage will parse a file with front-end YAML and markup content, and
// return a key-value Page structure.
func ParsePage(fn string) (Page, error) {
	c, err := ioutil.ReadFile(fn)
	if err != nil {
		return nil, err
	}
	return parsePage(fn, c)
}

// Helper function that creates a new Page from a byte array, parsing the
// front-end YAML and the markup, and pre-calculating all page-level variables.
func parsePage(fn string, c []byte) (Page, error) {

	page, err := parseMatter(c) //map[string] interface{} { }
	if err != nil {
		return nil, err
	}

	ext := filepath.Ext(fn)
	ext_output := ext
	markdown := isMarkdown(fn)

	// if markdown, change the output extension to html
	if markdown {
		ext_output = ".html"
	}

	page["ext"] = ext
	page["output_ext"] = ext_output
	page["id"] = removeExt(fn)
	page["url"] = replaceExt(fn, ext_output)

	// if markdown, convert to html
	raw := parseContent(c)
	if markdown {
		page["content"] = string(blackfriday.MarkdownCommon(raw))
	} else {
		page["content"] = string(raw)
	}

	if page["layout"] == nil {
		page["layout"] = "default"
	}

	// according to spec, Jekyll allows user to enter either category or
	// categories. Convert single category to string array to be consistent ...
	if category := page.GetString("category"); category != "" {
		page["categories"] = []string{category}
		delete(page, "category")
	}

	return page, nil
}

// Helper function to parse the front-end yaml matter.
func parseMatter(content []byte) (Page, error) {
	page := map[string]interface{}{}

	//first we extract the front YAML matter
	yamlMatter := bytes.Split(content, []byte("---\n"))[1]
	yamlParsed, err := goyaml2.Read(bytes.NewBuffer(yamlMatter))
	if err != nil {
		return nil, err
	}

	page = yamlParsed.(map[string]interface{})
	return page, err
}

// Helper function that separates the front-end yaml from the markup, and
// and returns only the markup (content) as a byte array.
func parseContent(content []byte) []byte {
	//now we need to parse out the markdown section create
	//buffered reader
	b := bytes.NewBuffer(content)
	m := new(bytes.Buffer)
	streams := 0

	//read each line of the file and read the markdown section
	//which is the second document stream in the yaml file
parse:
	for {
		line, err := b.ReadString('\n')
		switch {
		case err == io.EOF && streams >= 2:
			m.WriteString(line)
			break parse
		case err == io.EOF:
			break parse
		case err != nil:
			return nil
		case streams >= 2:
			m.WriteString(line)
		case strings.HasPrefix(line, "---"):
			streams++
		}
	}

	return m.Bytes()
}

// Sets a parameter value.
func (p Page) Set(key string, val interface{}) {
	p[key] = val
}

// Gets a parameter value.
func (p Page) Get(key string) interface{} {
	return p[key]
}

// Gets a parameter value as a string. If none exists return an empty string.
func (p Page) GetString(key string) (str string) {
	if v, ok := p[key]; ok {
		str = v.(string)
	}
	return
}

// Gets a parameter value as a string array.
func (p Page) GetStrings(key string) (strs []string) {
	if v, ok := p[key]; ok {
		for _, s := range v.([]interface{}) {
			strs = append(strs, s.(string))
		}
	}
	return
}

// Gets a parameter value as a byte array.
func (p Page) GetBytes(key string) (b []byte) {
	if v, ok := p[key]; ok {
		b = v.([]byte)
	}
	return
}

// Gets the layout file to use, without the extension.
// Layout files must be placed in the _layouts directory.
func (p Page) GetLayout() string {
	return p.GetString("layout")
}

// Gets the title of the Page.
func (p Page) GetTitle() string {
	return p.GetString("title")
}

// Gets the URL / relative path of the Page.
// e.g. /2008/12/14/my-post.html
func (p Page) GetUrl() string {
	return "/" + p.GetString("url")
}

// Gets the URL / relative path of the Page.
func (p Page) GetUrlRel() string {
	return p.GetString("url")
}

// Gets the Extension of the File (.html, .md, etc)
func (p Page) GetExt() string {
	return p.GetString("ext")
}

// Gets the un-rendered content of the Page.
func (p Page) GetContent() (c string) {
	if v, ok := p["content"]; ok {
		c = v.(string)
	}
	return
}

// Gets the list of tags to which this Post belongs.
func (p Page) GetTags() []string {
	return p.GetStrings("tags")
}

// Gets the of author to which this Post belongs.
func (p Page) GetAuthor() string {
	return p.GetString("author")
}

// Gets the of author to which this Post belongs.
func (p Page) GetAuthorLink() string {
	return p.GetString("authorlink")
}

// Gets the list of categories to which this post belongs.
func (p Page) GetCategories() []string {
	return p.GetStrings("categories")
}

func (p Page) ExtractFragment(regex, format string) (res string) {
	r := regexp.MustCompile(regex)
	src := ""
	res = ""
	matches := r.FindStringSubmatch(p["content"].(string))
	if matches != nil {
		src = matches[1]
		res = fmt.Sprintf(format, src)
	}
	return
}

func (p Page) MainVideoThumb() (src string) {
	return p.ExtractFragment(`embed/(.*?)"`, "http://img.youtube.com/vi/%s/default.jpg")
}

func (p Page) MainVideo() (src string) {
	return p.ExtractFragment(`<iframe (.*?)iframe>`, "<iframe %siframe>")
}

func (p Page) MainImg() (src string) {
	return p.ExtractFragment(`img src="(.*?)"`, "%s")
}

//Image Thumbnails
//http://api.imgur.com/models/image
func (p Page) MainImgSize(size string) (thumb_url string) {
	big_image := p.MainImg()
	if strings.Contains(big_image, "imgur") {
		r := regexp.MustCompile(`(.*imgur.*?)m\.([a-zA-Z]{3})$`)
		thumb_url = r.ReplaceAllString(big_image, "$1 "+size+".$2")
		//this shouldn't be necersary but the Replace all string was weird when using "$1s.$2"
		thumb_url = strings.Replace(thumb_url, " ", "", -1)
	} else {
		thumb_url = big_image
	}
	return
}

func (p Page) MainImgThumb() (thumb_url string) {
	return p.MainImgSize("s")
}

// Gets the list of categories to which this post belongs.
func (p Page) GetIntrotext() (intro string) {
	if _, found := p["introtext"]; found == true {
		intro = p.GetString("introtext")
	} else {
		intro = p["content"].(string)
		r, _ := regexp.Compile(`<img (.*?) />`) //eliminate the images,
		intro = r.ReplaceAllString(intro, "")
		r, _ = regexp.Compile(`<.*?>`) //eliminate the images,
		intro = r.ReplaceAllString(intro, "")
		r, _ = regexp.Compile(`</.*?>`) //eliminate the images,
		intro = r.ReplaceAllString(intro, "")

		max_chars := 300
		if len(intro) < max_chars {
			max_chars = len(intro)
		}
		intro = strings.TrimSpace(intro[:max_chars])
		intro = intro[0:strings.LastIndex(intro, " ")]
		intro = strings.Replace(intro, "\n\n", "</p>\n<p>", -1)
	}

	return
}
