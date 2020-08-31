package proxy

import (
	"net/http"
)

type requestPreparer func(incoming *http.Request, outgoing *http.Request) error
type RequestMiddleware func(next requestPreparer) requestPreparer

func FinishRequestPrep(incoming *http.Request, outgoing *http.Request) error {
	return nil
}

func CopyHeaders(next requestPreparer) requestPreparer {
	return func(i *http.Request, o *http.Request) error {
		o.Header = i.Header.Clone()
		return next(i, o)
	}
}

func CopyBody(next requestPreparer) requestPreparer {
	return func(i *http.Request, o *http.Request) error {
		o.Body = i.Body
		return next(i, o)
	}
}

func CopyContentLength(next requestPreparer) requestPreparer {
	return func(i *http.Request, o *http.Request) error {
		o.ContentLength = i.ContentLength
		return next(i, o)
	}
}
