package middleware

import (
	"net/http"
)

type RequestPreparer func(incoming *http.Request, outgoing *http.Request) error
type RequestMiddleware func(next RequestPreparer) RequestPreparer

func FinishRequestPrep(incoming *http.Request, outgoing *http.Request) error {
	return nil
}

func CopyHeaders(next RequestPreparer) RequestPreparer {
	return func(i *http.Request, o *http.Request) error {
		o.Header = i.Header.Clone()
		return next(i, o)
	}
}

func CopyBody(next RequestPreparer) RequestPreparer {
	return func(i *http.Request, o *http.Request) error {
		o.Body = i.Body
		return next(i, o)
	}
}

func CopyContentLength(next RequestPreparer) RequestPreparer {
	return func(i *http.Request, o *http.Request) error {
		o.ContentLength = i.ContentLength
		return next(i, o)
	}
}
