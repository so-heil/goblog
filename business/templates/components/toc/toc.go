package toc

import "strings"

func ElementID(title string) string {
	return strings.Replace(title, " ", "-", -1)
}
