# reverse-proxy

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

#### Request Middleware

Coming soon!

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
