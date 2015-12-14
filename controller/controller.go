package controller

// this file implements helper functions
// for controllers in the server to use

import (
	"net/http"
	"strings"
)

// function path components returns the path components
// for the url in a given request
func pathComponents(r *http.Request) (comps []string) {
	comps = strings.Split(r.URL.Path, "/")
	return
}

// function rid returns the id part of the request
func rid(r *http.Request) string {
	comps := pathComponents(r)
	return comps[0]
}

func urlSubcategory(r *http.Request) string {
	comps := pathComponents(r)
	if len(comps) > 1 {
		return comps[1]
	}
	return ""
}
