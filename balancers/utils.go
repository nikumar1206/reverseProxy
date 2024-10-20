package balancers

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
