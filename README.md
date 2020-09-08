# reverse-proxy

![Tests](https://github.com/jdlubrano/reverse-proxy/workflows/Tests/badge.svg)

## Problem

For organizations that choose to host their services on an "internal" network,
behind a VPN or on a VPC or whatever, use cases can arise such that you want
an internal service to be able to serve specific requests from the public
internet via a specific API endpoint (over HTTP).  A common example is that you
wish to accept a webhook from a trusted third-party.  How can you solve this
problem?

You can definitely use tools like AWS API Gateway to control access to private
APIs.  If you can use API Gateway to meet your goals, that is a great option.
API Gateway has _a lot_ of functionality that you may not need, however.  It
has a learning curve that you may not be willing to climb to accomplish
something simple or something intended to be a proof-of-concept or temporary.
Maybe your dev organization does not allow all developers to have unrestricted
AWS access (for good reason).  In any of those exceptions, this Go web service
might provide a simple alternative.

## Solution

This project is meant to be a turnkey solution to start routing requests from
the public internet to endpoints on a private network.  The intended use is to
set up a 1-to-1 mapping of endpoints in this
[reverse proxy](https://www.cloudflare.com/learning/cdn/glossary/reverse-proxy/)
so that requests can be forwarded to internal services.  This reverse proxy
will need access to your internal services and it will need to be available
to the public internet.  The idea is that it would be more secure to have this
reverse proxy on the public internet than to expose every individual service
that needs to receive a webhook to be exposed on the public internet.

## Usage

### Docker Quick Start

You should be able to run a version of this reverse proxy without writing any
Go code.

Let's start with a new directory:

```
mkdir my-reverse-proxy
cd my-reverse-proxy
```

Now create a routes file at `routes.yml` and open that file in an editor.
Refer to the example routes file below and modify it based on your needs.

```yaml
routes:
  - incoming_request_path: '/hello'
    forwarded_request_url: 'https://my.internal.service.com'
    forwarded_request_path: '/my/internal/api'
```

Given this configuration, any requests made to the reverse proxy with a path
of `/hello` will be forwarded, headers, query, request body, et. all to your
internal endpoint, `https://my.internal.service.com/my/internal/api`.  You can
specify as many routes as you like.

Now, once you are finished configuring your routes, create a Dockerfile.

```
FROM jdlubrano:reverse-proxy

COPY routes.yml ./routes.yml

ENV REVERSE_PROXY_PORT 8080
ENV REVERSE_PROXY_ROUTES_FILE routes.yml

CMD ["reverse-proxy"]
```

The `REVERSE_PROXY_PORT` and `REVERSE_PROXY_ROUTES_FILE` environment variables
should be changed based on your requirements.  You should be able to build the
Docker image from that Dockerfile and `docker run` the reverse proxy.

### Customizing Behavior

The `main.go` file in this repo attempts to be a barebones example of how to
start the `Proxy` within a Go app.  Some of the features of the reverse proxy
are not used in `main.go`, however.

#### Custom Handlers

If you need to add an endpoint to your proxy, you can add a custom `http.Handler`.
This can be especially helpful if you need to configure a healthcheck endpoint
or something of that nature.

```go
import (
        "net/http"

	"github.com/jdlubrano/reverse-proxy/internal/proxy"
)

handlerFunc := func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(200)
        w.Write([]byte(`{"status": "OK"}`))
}

proxy := proxy.NewProxy(...)
proxy.AddCustomHandler("/healthcheck", http.HandlerFunc(handlerFunc))
// ...start the proxy
```

#### Request Middleware

Request middleware is designed to modify the outgoing request that eventually
gets forwarded to the configured route/destination.  There are two types to
be aware of when it comes to implementing your own request middleware.

The first type is a `RequestPreparer`.  A `RequestPreparer` is defined as a

```go
type RequestPreparer func(incoming *http.Request, outgoing *http.Request) error
```

The `incoming` request is the request received by the proxy.  The `outgoing`
request is the request that will eventually be forwarded a downstream location
(as configured by the routes).

You will most likely not define a `RequestPreparer`, however.  You are far
more likely to define a `RequestMiddleware`.  A `RequestMiddleware` is defined
as

```go
type RequestMiddleware func(next RequestPreparer) RequestPreparer
```

By accepting a `next` parameter, you can easily make your middleware integrate
with existing `RequestMiddleware`.

For example, you could add a header to the outgoing request:

```go
func AddMyHeader(next middleware.RequestPreparer) middleware.RequestPreparer {
        return func(incoming *http.Request, outgoing *http.Request) error {
                outgoing.Header.Add("My-Header", "My Content")
                return next(incoming, outgoing)
        }
}
```

You can insert your `RequestMiddleware` anywhere in the `RequestMiddleware`
chain of your `Proxy`.

```go
import (
        "net/http"

	"github.com/jdlubrano/reverse-proxy/internal/middleware"
	"github.com/jdlubrano/reverse-proxy/internal/proxy"
)

proxy := proxy.NewProxy(...)

// Adding middleware to the end of the middleware chain
proxy.RequestMiddleware = append(proxy.RequestMiddleware, AddMyHeader)

// Adding middleware to the start of the middleware chain.
// Note that the default middleware would run after your custom middleware in
// this case.  The default middleware may undo whatever your middlware is doing.
proxy.RequestMiddleware = append([]middleware.RequestMiddleware{AddMyHeader}, ...proxy.RequestMiddleware)

// ...start the proxy
```

`RequestMiddleware` at the end of the chain runs after `RequestMiddleware` at
the beginning of the chain.  The default middleware chain does three things
that are most likely desired behavior for a reverse proxy.  Namely the default
request middleware:

1. Copies request headers from `incoming` to `outgoing`. (`CopyHeaders`)
2. Copies the request body from `incoming` to `outgoing`. (`CopyBody`)
3. Fills the `Content-Length` header of `outgoing`. (`CopyContentLength`)

#### Roundtrip Middleware

Coming soon!

## Requirements

* Go >= 1.14

## Development

Build the project
```
go build
```

Start the Reverse Proxy
```
./reverse-proxy
```
