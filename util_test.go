package main

import (
	"testing"
)

func TestAppendExt(t *testing.T) {
	if ext := appendExt("/test.html", ".html"); ext != "/test.html" {
		t.Errorf("Expected appended extension [/test.html] got [%s]", ext)
	}
	if ext := appendExt("/test", ".html"); ext != "/test.html" {
		t.Errorf("Expected appended extension [/test.html] got [%s]", ext)
	}
}

func TestHasMatter(t *testing.T) {
	// TODO
}

func TestIsHiddenOrTemp(t *testing.T) {
	tests := map[string]bool{
		".tmp": true,
		"tmp~": true,
		"tmp":  false,
		".git": true}

	for key, val := range tests {
		if result := isHiddenOrTemp(key); result != val {
			t.Errorf("Expected IsHiddenOrTemp value of [%v] got [%v] for file [%s]", val, result, key)
		}
	}
}

func TestIsTemplate(t *testing.T) {
	tests := map[string]bool{
		"_layouts/page.html":   true,
		"_includes/page.html":  true,
		"_includes/page.html~": false,
		"static/js/script.js":  false,
		"index.html":           false}

	for key, val := range tests {
		if result := isTemplate(key); result != val {
			t.Errorf("Expected IsTemplate value of [%v] got [%v] for file [%s]", val, result, key)
		}
	}
}

func TestIsHtml(t *testing.T) {
	tests := map[string]bool{
		"page.html":  true,
		"page.xml":   true,
		"page.html~": false,
		"page.rss":   true,
		"page.atom":  true}

	for key, val := range tests {
		if result := isHtml(key); result != val {
			t.Errorf("Expected IsHtml value of [%v] got [%v] for file [%s]", val, result, key)
		}
	}
}

func TestIsMarkdown(t *testing.T) {
	tests := map[string]bool{
		"page.md":       true,
		"page.markdown": true,
		"page.md~":      false}

	for key, val := range tests {
		if result := isMarkdown(key); result != val {
			t.Errorf("Expected IsMarkdown value of [%v] got [%v] for file [%s]", val, result, key)
		}
	}
}

func TestIsPage(t *testing.T) {
	// TODO
}

func TestIsPost(t *testing.T) {
	// TODO
}

func TestIsStatic(t *testing.T) {
	tests := map[string]bool{
		"_site":            false,
		"_site/index.html": false,
		"img/logo.png":     true}

	for key, val := range tests {
		if result := isStatic(key); result != val {
			t.Errorf("Expected isStatic value of [%v] got [%v] for file [%s]", val, result, key)
		}
	}
}

func TestRemoveExt(t *testing.T) {
	if ext := removeExt("/test"); ext != "/test" {
		t.Errorf("Expected removed extension [/test] got [%s]", ext)
	}
	if ext := removeExt("/test.html"); ext != "/test" {
		t.Errorf("Expected removed extension [/test] got [%s]", ext)
	}
}

func TestReplaceExt(t *testing.T) {
	if ext := replaceExt("/test", ".html"); ext != "/test.html" {
		t.Errorf("Expected replaced extension [/test.html] got [%s]", ext)
	}
	if ext := replaceExt("/test.markdown", ".html"); ext != "/test.html" {
		t.Errorf("Expected replaced extension [/test.html] got [%s]", ext)
	}
}
