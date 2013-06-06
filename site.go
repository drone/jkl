package main

import (
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
	Src       string // Directory where Jekyll will look to transform files
	Dest      string // Directory where Jekyll will write files to
	Conf      Config // Configuration date from the _config.yml file
	Something string

	posts []Page // Posts thet need to be generated
	pages []Page // Pages that need to be generated

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

	auth := aws.Auth{user, pass}
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

	// Lists of templates (_layouts, _includes) that we find thate
	// will need to be compiled
	layouts := []string{}

	// func to walk the jekyll directory structure
	walker := func(fn string, fi os.FileInfo, err error) error {

		rel, _ := filepath.Rel(s.Src, fn)
		switch {
		case err != nil:
			return nil

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
			if post["published"] == true {
				// TODO: this is a hack to get the posts in rev chronological order
				s.posts = append([]Page{post}, s.posts...) //s.posts, post)
			}

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

	// Compile all templates found
	//s.templ = template.Must(template.ParseFiles(layouts...))
	s.templ, err = template.New("layouts").Funcs(funcMap).ParseFiles(layouts...)
	if err != nil {
		return err
	}

	// Add the posts, timestamp, etc to the Site Params
	s.Conf.Set("posts", s.posts)
	s.Conf.Set("time", time.Now())

	s.calculateTags()
	s.calculateCategories()
	s.SetMinuteByMinute()
	s.calculateAuthors()
	s.getPostByAuthor()
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

		// make sure the posts's parent dir exists
		d := filepath.Join(s.Dest, filepath.Dir(url))
		f := filepath.Join(s.Dest, url)
		if err := os.MkdirAll(d, 0755); err != nil {
			return err
		}

		//data passed in to each template
		data := map[string]interface{}{
			"s":    s,
			"site": s.Conf,
			"page": page,
		}

		buf, err := page.RenderTemplate(s, data)
		if err != nil {
			return err
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

// Helper function to aggregate a list of all categories anxod their
// related posts.
func (s *Site) calculateCategories() {

	categories := make(map[string][]Page)
	pages := []Page{}
	pages = append(pages, s.pages...)
	pages = append(pages, s.posts...)

	//Assuming that posts is sorted from most recent to least recent.
	max_post := 1200
	if len(pages) < max_post {
		max_post = len(pages)
	}

	latest_pages := pages[:max_post]
	for _, page := range latest_pages {
		for _, category := range page.GetCategories() {
			if posts, ok := categories[category]; ok == true {
				categories[category] = append(posts, page)
			} else {
				categories[category] = []Page{page}
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
			if posts, ok := tags[tag]; ok == true {
				tags[tag] = append(posts, post)
			} else {
				tags[tag] = []Page{post}
			}
		}
	}

	s.Conf.Set("tags", tags)
}

func (s *Site) calculateAuthors() {

	authors := make(map[string]string)
	for _, post := range s.posts {
		author := post.GetAuthor()
		authors[author] = post.GetAuthorLink()
	}
	s.Conf.Set("authors", authors)
}

func getSubStr(word string, arr []string) bool {
	flag := false
	for i := 0; i < len(arr); i++ {
		if word == arr[i] {
			flag = true
		}
	}
	return flag
}

func getSubArr(word []string, arr []string) bool {
	flag := false
	for i := 0; i < len(arr); i++ {
		for j := 0; j < len(word); j++ {
			if word[j] == arr[i] {
				flag = true
			}
		}
	}
	return flag
}

func (s *Site) SetMinuteByMinute() {
	max_post := 60
	minbymin := []string{"Autor", "Publicidad","Publicidad2","Publicidad3" }
	min_posts := []Page{}
	if len(s.posts) < max_post {
		max_post = len(s.posts)
	}
	latest_posts := s.posts[:max_post]
	for _, post := range latest_posts {

		if !getSubArr(post.GetCategories(), minbymin) {
			min_posts = append(min_posts, post)
		}

	}

	s.Conf.Set("MinuteByMinute", min_posts)
}

func (s *Site) getPostByAuthor() {
	autor_posts := make(map[string][]Page)

	for _, post := range s.posts {
		if posts, ok := autor_posts[post.GetAuthor()]; ok == true {
			autor_posts[post.GetAuthor()] = append(posts, post)
		} else {
			autor_posts[post.GetAuthor()] = []Page{post}
		}
	}

	small_autor := make(map[string][]Page)

	for autor_name, autor_news := range autor_posts {
		small_autor[autor_name] = cutArr(autor_news, 30)
	}

	s.Conf.Set("postByAuthor", small_autor)
}

func (s *Site) PagesForCategory(cat string, max_pages int) []Page {
	categories := s.Conf.Get("categories").(map[string][]Page)
	return cutArr(categories[cat], max_pages)
}

func cutArr(news []Page, max_post int) []Page {
	if len(news) < max_post {
		max_post = len(news)
	}
	return news[:max_post]

}
