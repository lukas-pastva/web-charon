package handlers

import (
	"regexp"
	"strings"
)

// Czech and common diacritics → ASCII mapping.
var diacriticsReplacer = strings.NewReplacer(
	"á", "a", "Á", "a",
	"č", "c", "Č", "c",
	"ď", "d", "Ď", "d",
	"é", "e", "É", "e",
	"ě", "e", "Ě", "e",
	"í", "i", "Í", "i",
	"ň", "n", "Ň", "n",
	"ó", "o", "Ó", "o",
	"ř", "r", "Ř", "r",
	"š", "s", "Š", "s",
	"ť", "t", "Ť", "t",
	"ú", "u", "Ú", "u",
	"ů", "u", "Ů", "u",
	"ý", "y", "Ý", "y",
	"ž", "z", "Ž", "z",
)

var (
	slugNonAlnum = regexp.MustCompile(`[^a-z0-9]+`)
	slugTrim     = regexp.MustCompile(`^-|-$`)
)

// Slugify converts a string (including Czech diacritics) into a URL-friendly slug.
func Slugify(s string) string {
	s = strings.ToLower(s)
	s = diacriticsReplacer.Replace(s)
	s = slugNonAlnum.ReplaceAllString(s, "-")
	s = slugTrim.ReplaceAllString(s, "")
	return s
}
