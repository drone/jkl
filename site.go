package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"launchpad.net/goamz/aws"
	"launchpad.net/goamz/s3"
	"mime"
	"os"
	"path/filepath"
	"text/template"
	"time"
)

var (
	MsgCopyingFile  = "Copying File: %s"
	MsgGenerateFile = "Generating Page: %s"
	MsgUploadFile   = "Uploading: %s"
	MsgUsingConfig  = "Loading Config: %s"
)

type Site struct {
	Src  string // Directory where Jekyll will look to transform files
	Dest string // Directory where Jekyll will write files to
	Conf Config // Configuration date from the _config.yml file

	posts []Page             // Posts thet need to be generated
	pages []Page             // Pages that need to be generated
	files []string           // Static files to get copied to the destination
	templ *template.Template // Compiled templates
}

func NewSite(src, dest string) (*Site, error) {

	// Parse the _config.yml file
	path := filepath.Join(src, "_config.yml")
	conf, err := ParseConfig(path)
	logf(MsgUsingConfig, path)
	if err != nil {
		return nil, err
	}

	site := Site{
		Src:  src,
		Dest: dest,
		Conf: conf,
	}

	// Recursively process all files in the source directory
	// and parse pages, posts, templates, etc
	if err := site.read(); err != nil {
		return nil, err
	}

	return &site, nil
}

// Reloads the site into memory
func (s *Site) Reload() error {
	s.posts = []Page{}
	s.pages = []Page{}
	s.files = []string{}
	s.templ = nil
	return s.read()
}

// Prepares the source directory for site generation
func (s *Site) Prep() error {
	return os.MkdirAll(s.Dest, 0755)
}

// Removes the existing site (typically in _site).
func (s *Site) Clear() error {
	return os.RemoveAll(s.Dest)
}

// Generates a static website based on Jekyll standard layout.
func (s *Site) Generate() error {

	// Remove previously generated site, and then (re)create the
	// destination directory
	if err := s.Clear(); err != nil {
		return err
	}
	if err := s.Prep(); err != nil {
		return err
	}

	// Generate all Pages and Posts and static files
	if err := s.writePages(); err != nil {
		return err
	}
	if err := s.writeStatic(); err != nil {
		return err
	}

	return nil
}

// Deploys a site to S3.
func (s *Site) Deploy(user, pass, url string) error {

	auth := aws.Auth{AccessKey: user, SecretKey: pass}
	b := s3.New(auth, aws.USEast).Bucket(url)

	// walks _site directory and uploads file to S3
	walker := func(fn string, fi os.FileInfo, err error) error {
		if fi.IsDir() {
			return nil
		}

		rel, _ := filepath.Rel(s.Dest, fn)
		typ := mime.TypeByExtension(filepath.Ext(rel))
		content, err := ioutil.ReadFile(fn)
		logf(MsgUploadFile, rel)
		if err != nil {
			return err
		}

		// try to upload the file ... sometimes this fails due to amazon
		// issues. If so, we'll re-try
		if err := b.Put(rel, content, typ, s3.PublicRead); err != nil {
			time.Sleep(100 * time.Millisecond) // sleep so that we don't immediately retry
			return b.Put(rel, content, typ, s3.PublicRead)
		}

		// file upload was a success, return nil
		return nil
	}

	return filepath.Walk(s.Dest, walker)
}

// Helper function to traverse the source directory and identify all posts,
// projects, templates, etc and parse.
func (s *Site) read() error {

	// Lists of templates (_layouts, _includes) that we find that
	// will need to be compiled
	layouts := []string{}

	// func to walk the jekyll directory structure
	walker := func(fn string, fi os.FileInfo, err error) error {
		rel, _ := filepath.Rel(s.Src, fn)
		switch {
		case err != nil:
			return nil

		case fi.IsDir() && isHiddenOrTemp(fn):
			return filepath.SkipDir

		// Ignore directories
		case fi.IsDir():
			return nil

		// Ignore Hidden or Temp files
		// (starting with . or ending with ~)
		case isHiddenOrTemp(rel):
			return nil

		// Parse Templates
		case isTemplate(rel):
			layouts = append(layouts, fn)

		// Parse Posts
		case isPost(rel):
			post, err := ParsePost(rel)
			if err != nil {
				return err
			}
			// TODO: this is a hack to get the posts in rev chronological order
			s.posts = append([]Page{post}, s.posts...) //s.posts, post)

		// Parse Pages
		case isPage(rel):
			page, err := ParsePage(rel)
			if err != nil {
				return err
			}
			s.pages = append(s.pages, page)

		// Move static files, no processing required
		case isStatic(rel):
			s.files = append(s.files, rel)
		}
		return nil
	}

	// Walk the diretory recursively to get a list of all posts,
	// pages, templates and static files.
	err := filepath.Walk(s.Src, walker)
	if err != nil {
		return err
	}

	// Compile all templates found, if any
	if len(layouts) > 0 {
		s.templ, err = template.New("layouts").Funcs(funcMap).ParseFiles(layouts...)
		if err != nil {
			return err
		}
	}

	// Add the posts, timestamp, etc to the Site Params
	s.Conf.Set("posts", s.posts)
	s.Conf.Set("time", time.Now())
	s.calculateTags()
	s.calculateCategories()

	return nil
}

// Helper function to write all pages and posts to the destination directory
// during site generation.
func (s *Site) writePages() error {

	// There is really no difference between a Page and a Post (other than
	// initial parsing) so we can combine the lists and use the same rendering
	// code for both.
	pages := []Page{}
	pages = append(pages, s.pages...)
	pages = append(pages, s.posts...)

	for _, page := range pages {
		url := page.GetUrl()
		layout := page.GetLayout()

		// is the layout provided? or is it nil /empty?
		//layoutNil := layout == "" || layout == "nil"

		// make sure the posts's parent dir exists
		d := filepath.Join(s.Dest, filepath.Dir(url))
		f := filepath.Join(s.Dest, url)
		if err := os.MkdirAll(d, 0755); err != nil {
			return err
		}

		// if markdown, need to convert to html
		// otherwise just convert raw html to a string
		//var content string
		//if isMarkdown(page.GetExt()) {
		//	content = string(blackfriday.MarkdownCommon(raw))
		//} else {
		//	content = string(raw)
		//}

		//data passed in to each template
		data := map[string]interface{}{
			"site": s.Conf,
			"page": page,
		}

		// treat all non-markdown pages as templates
		content := page.GetContent()
		if isMarkdown(page.GetExt()) == false {
			// this code will add the page to the list of templates,
			// will execute the template, and then set the content
			// to the rendered template

			if s.templ == nil {
				return fmt.Errorf("No templates defined for page: %s", url)
			}

			t, err := s.templ.New(url).Parse(content)
			if err != nil {
				return err
			}
			var buf bytes.Buffer
			err = t.ExecuteTemplate(&buf, url, data)
			if err != nil {
				return err
			}
			content = buf.String()
		}

		// add document body to the map
		data["content"] = content

		// write the template to a buffer
		// NOTE: if template is nil or empty, then we should parse the
		//       content as if it were a template
		var buf bytes.Buffer
		if layout == "" || layout == "nil" {
			//t, err := s.templ.New(url).Parse(content);
			//if err != nil { return err }
			//err = t.ExecuteTemplate(&buf, url, data);
			//if err != nil { return err }

			buf.WriteString(content)
		} else {
			layout = appendExt(layout, ".html")
			s.templ.ExecuteTemplate(&buf, layout, data)
		}

		logf(MsgGenerateFile, url)
		if err := ioutil.WriteFile(f, buf.Bytes(), 0644); err != nil {
			return err
		}
	}

	return nil
}

// Helper function to write all static files to the destination directory
// during site generation. This will also take care of creating any parent
// directories, if necessary.
func (s *Site) writeStatic() error {

	for _, file := range s.files {
		from := filepath.Join(s.Src, file)
		to := filepath.Join(s.Dest, file)
		logf(MsgCopyingFile, file)
		if err := copyTo(from, to); err != nil {
			return err
		}
	}

	return nil
}

// Helper function to aggregate a list of all categories and their
// related posts.
func (s *Site) calculateCategories() {

	categories := make(map[string][]Page)
	for _, post := range s.posts {
		for _, category := range post.GetCategories() {
			if posts, ok := categories[category]; ok == false {
				categories[category] = append(posts, post)
			} else {
				categories[category] = []Page{post}
			}
		}
	}

	s.Conf.Set("categories", categories)
}

// Helper function to aggregate a list of all tags and their
// related posts.
func (s *Site) calculateTags() {

	tags := make(map[string][]Page)
	for _, post := range s.posts {
		for _, tag := range post.GetTags() {
			if posts, ok := tags[tag]; ok == false {
				tags[tag] = append(posts, post)
			} else {
				tags[tag] = []Page{post}
			}
		}
	}

	s.Conf.Set("tags", tags)
}
