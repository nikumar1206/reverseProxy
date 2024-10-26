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

func filter[S ~[]E, E any](s S, cmp func(a E) bool) S {
	var finalArray []E
	for _, ele := range s {
		if cmp(ele) {
			finalArray = append(finalArray, ele)
		}
	}
	return finalArray
}
