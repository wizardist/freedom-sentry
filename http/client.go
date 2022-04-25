package http

import gohttp "net/http"

type Client interface {
	// Do sends an HTTP request and returns an HTTP response, following
	// policy (such as redirects, cookies, auth) as configured on the
	// client.
	//
	// An error is returned if caused by client policy (such as
	// CheckRedirect), or failure to speak HTTP (such as a network
	// connectivity problem). A non-2xx status code doesn't cause an
	// error.
	//
	// If the returned error is nil, the Response will contain a non-nil
	// Body which the user is expected to close. If the Body is not both
	// read to EOF and closed, the Client's underlying RoundTripper
	// (typically Transport) may not be able to re-use a persistent TCP
	// connection to the server for a subsequent "keep-alive" request.
	//
	// The request Body, if non-nil, will be closed by the underlying
	// Transport, even on errors.
	//
	// On error, any Response can be ignored. A non-nil Response with a
	// non-nil error only occurs when CheckRedirect fails, and even then
	// the returned Response.Body is already closed.
	//
	// Generally Get, Post, or PostForm will be used instead of Do.
	//
	// If the server replies with a redirect, the Client first uses the
	// CheckRedirect function to determine whether the redirect should be
	// followed. If permitted, a 301, 302, or 303 redirect causes
	// subsequent requests to use HTTP method GET
	// (or HEAD if the original request was HEAD), with no body.
	// A 307 or 308 redirect preserves the original HTTP method and body,
	// provided that the Request.GetBody function is defined.
	// The NewRequest function automatically sets GetBody for common
	// standard library body types.
	//
	// Any returned error will be of type *url.Error. The url.Error
	// value's Timeout method will report true if the request timed out.
	Do(req *gohttp.Request) (*gohttp.Response, error)
}
