package checker

import "strings"

func joinURL(baseURL string, path string) string {
	baseURL = strings.TrimRight(baseURL, "/")
	path = strings.TrimLeft(path, "/")

	return baseURL + "/" + path
}
