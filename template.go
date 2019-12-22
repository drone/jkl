package main

import (
	"strings"
	"time"
)

// Additional functions available in Jekyll templates
var funcMap = map[string]interface{}{

	"capitalize":        capitalize,
	"date_to_string":    dateToString,
	"date_to_xmlschema": dateToXmlSchema,
	"downcase":          lower,
	"eq":                eq,
	"newline_to_br":     newlineToBreak,
	"replace":           replace,
	"replace_first":     replaceFirst,
	"remove":            remove,
	"remove_first":      removeFirst,
	"split":             split,
	"strip_newlines":    stripNewlines,
	"truncate":          truncate,
	"truncatewords":     truncateWords,
	"upcase":            upper,
}

// Capitalize words in the input sentence
func capitalize(s string) string {
	return strings.Title(s)
}

// Checks if two values are equal
func eq(v1 interface{}, v2 interface{}) bool {
	return v1 == v2
}

// Converts a date to a string
func dateToString(date time.Time) string {
	return date.Format("2006-01-02")
}

// Converts a date to a string
func dateToXmlSchema(date time.Time) string {
	return date.Format(time.RFC3339)
}

// Convert an input string to lowercase
func lower(s string) string {
	return strings.ToLower(s)
}

// Replace each newline (\n) with html break
func newlineToBreak(s string) string {
	return strings.Replace(s, "\n", "<br/>", -1)
}

// Remove each occurrence
func remove(s, pattern string) string {
	return strings.Replace(s, pattern, "", -1)
}

// Remove the first occurrence
func removeFirst(s, pattern string) string {
	return strings.Replace(s, pattern, "", 1)
}

// Replace each occurrence
func replace(s, old, new string) string {
	return strings.Replace(s, old, new, -1)
}

// Replace the first occurrence
func replaceFirst(s, old, new string) string {
	return strings.Replace(s, old, new, 1)
}

// Split a string on a matching pattern
func split(s, pattern string) []string {
	return strings.Split(s, pattern)
}

// Strip all newlines (\n) from string
func stripNewlines(s string) string {
	return strings.Replace(s, "\n", "", -1)
}

// Truncate a string down to x characters
func truncate(s string, x int) string {
	if len(s) > x {
		return s[0:x]
	}
	return s
}

// Truncate a string down to x words
func truncateWords(s string, x int) string {
	words := strings.Split(s, " ")
	if len(words) <= x {
		return s
	}
	return strings.Join(words[0:x], " ")
}

// Convert an input string to uppercase
func upper(s string) string {
	return strings.ToUpper(s)
}
