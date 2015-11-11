package url

import "strings"

// Join joins any number of path elements into a single path, adding a separating slash if necessary.
// All empty strings are ignored.
func Join(elem ...string) string {
	for i, e := range elem {
		if e != "" {
			return clean(strings.Join(elem[i:], "/"))
		}
	}
	return ""
}

func clean(urlPart string) string {
	return strings.TrimRight(urlPart, "/")
}
