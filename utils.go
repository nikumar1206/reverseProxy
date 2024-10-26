package main

import (
	"net/http"
	"net/url"
)

func updateRequest(req *http.Request, newURL url.URL) *http.Request {
	req.URL.Host = newURL.Host
	req.URL.Scheme = newURL.Scheme
	req.RequestURI = ""
	return req
}

// Filter returns a new slice that satisfies the conditions in the cmp func.
func Filter[S ~[]E, E any](s S, cmp func(a E) bool) S {
	var finalArray []E
	for _, ele := range s {
		if cmp(ele) {
			finalArray = append(finalArray, ele)
		}
	}
	return finalArray
}
