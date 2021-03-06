package proxy

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jdlubrano/reverse-proxy/internal/logger"
	"github.com/jdlubrano/reverse-proxy/internal/middleware"
	"github.com/jdlubrano/reverse-proxy/internal/routes"
	"github.com/stretchr/testify/assert"
)

func TestProxy(t *testing.T) {
	downstreamServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/error" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(500)
			w.Write([]byte(`{"error": "Something went wrong"}`))
		}

		if r.URL.Path == "/success" {
			w.Write([]byte(`{"success": true}`))
		}

		if r.URL.Path == "/headers" {
			w.Header().Set("X-Added-Downstream", "header added by downstream server")
			w.Header().Set("X-Received-Downstream", r.Header.Get("X-Sent-Downstream"))
			w.Write([]byte(`{"success": true}`))
		}

		if r.URL.Path == "/query" {
			response := fmt.Sprintf(`{"received": "%s"}`, r.URL.Query().Get("sent"))

			w.WriteHeader(200)
			w.Write([]byte(response))
		}

		if r.URL.Path == "/body" {
			bodyParams := make(map[string]string)
			err := json.NewDecoder(r.Body).Decode(&bodyParams)

			if err != nil {
				w.WriteHeader(500)
				fmt.Println("JSON decoding error", err)
				return
			}

			response := fmt.Sprintf(`{"received": "%s"}`, bodyParams["sent"])

			w.WriteHeader(200)
			w.Write([]byte(response))
		}
	}))

	defer downstreamServer.Close()

	errorRoute := routes.Route{
		IncomingRequestPath:  "/test/error",
		ForwardedRequestURL:  downstreamServer.URL,
		ForwardedRequestPath: "/error",
	}

	successRoute := routes.Route{
		IncomingRequestPath:  "/test/success",
		ForwardedRequestURL:  downstreamServer.URL,
		ForwardedRequestPath: "/success",
	}

	headersRoute := routes.Route{
		IncomingRequestPath:  "/test/headers",
		ForwardedRequestURL:  downstreamServer.URL,
		ForwardedRequestPath: "/headers",
	}

	bodyRoute := routes.Route{
		IncomingRequestPath:  "/test/body",
		ForwardedRequestURL:  downstreamServer.URL,
		ForwardedRequestPath: "/body",
	}

	queryRoute := routes.Route{
		IncomingRequestPath:  "/test/query",
		ForwardedRequestURL:  downstreamServer.URL,
		ForwardedRequestPath: "/query",
	}

	routesConfig := &routes.RoutesConfig{
		Routes: []routes.Route{
			successRoute,
			errorRoute,
			headersRoute,
			bodyRoute,
			queryRoute,
		},
	}

	logger := &logger.NullLogger{}
	proxy := NewProxy(logger, routesConfig, 8080)

	go func() {
		proxy.Start()
	}()
	defer proxy.Stop()

	assert := assert.New(t)

	t.Run("when incoming request does not match any Routes", func(t *testing.T) {
		resp, err := http.Get("http://localhost:8080/not_found")
		defer resp.Body.Close()

		assert.Nil(err)
		assert.Equal(404, resp.StatusCode)
	})

	t.Run("when incoming request does not quite match any Routes", func(t *testing.T) {
		resp, err := http.Get("http://localhost:8080/test/error/specific_error")
		defer resp.Body.Close()

		assert.Nil(err)
		assert.Equal(404, resp.StatusCode)
	})

	t.Run("when the downstream endpoint returns an error", func(t *testing.T) {
		resp, err := http.Get("http://localhost:8080/test/error")
		defer resp.Body.Close()
		assert.Nil(err)

		body, err := ioutil.ReadAll(resp.Body)
		assert.Nil(err)

		assert.Equal(500, resp.StatusCode)
		assert.Equal(string(body), `{"error": "Something went wrong"}`)
	})

	t.Run("when the downstream endpoint returns a successful response", func(t *testing.T) {
		resp, err := http.Get("http://localhost:8080/test/success")
		assert.Nil(err)
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		assert.Nil(err)

		assert.Equal(200, resp.StatusCode)
		assert.Equal(string(body), `{"success": true}`)
	})

	t.Run("when headers are provided", func(t *testing.T) {
		req, err := http.NewRequest("GET", "http://localhost:8080/test/headers", nil)
		assert.Nil(err)

		req.Header.Add("X-Sent-Downstream", "test sent header")
		resp, err := http.DefaultClient.Do(req)
		assert.Nil(err)

		assert.Equal(200, resp.StatusCode)
		assert.Equal("test sent header", resp.Header.Get("X-Received-Downstream"))
	})

	t.Run("when headers are added by the downstream endpoint", func(t *testing.T) {
		resp, err := http.Get("http://localhost:8080/test/headers")
		assert.Nil(err)
		defer resp.Body.Close()

		assert.Equal(200, resp.StatusCode)
		assert.Equal("header added by downstream server", resp.Header.Get("X-Added-Downstream"))
	})

	t.Run("when a request query string is provided", func(t *testing.T) {
		resp, err := http.Get("http://localhost:8080/test/query?sent=test")
		assert.Nil(err)
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		assert.Nil(err)

		assert.Equal(200, resp.StatusCode)
		assert.Equal(string(body), `{"received": "test"}`)
	})

	t.Run("when a request body is provided", func(t *testing.T) {
		requestBody := `{"sent": "test"}`
		resp, err := http.Post("http://localhost:8080/test/body", "application/json", strings.NewReader(requestBody))
		assert.Nil(err)
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		assert.Nil(err)

		assert.Equal(200, resp.StatusCode)
		assert.Equal(string(body), `{"received": "test"}`)
	})
}

func TestProxyCustomHandler(t *testing.T) {
	routesConfig := &routes.RoutesConfig{}
	logger := &logger.NullLogger{}
	proxy := NewProxy(logger, routesConfig, 8080)

	handlerFunc := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"status": "OK"}`))
	}

	proxy.AddCustomHandler("/healthcheck", http.HandlerFunc(handlerFunc))

	go func() {
		proxy.Start()
	}()
	defer proxy.Stop()

	assert := assert.New(t)
	resp, err := http.Get("http://localhost:8080/healthcheck")

	assert.Nil(err)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	assert.Nil(err)
	assert.Equal(200, resp.StatusCode)
	assert.Equal(`{"status": "OK"}`, string(body))
}

func TestProxyRequestMiddleware(t *testing.T) {
	assert := assert.New(t)
	logger := &logger.NullLogger{}
	headerPresent := false

	addMyHeader := func(next middleware.RequestPreparer) middleware.RequestPreparer {
		return func(i *http.Request, o *http.Request) error {
			o.Header.Add("X-MyHeader", "test")
			return next(i, o)
		}
	}

	downstreamServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headerPresent = r.Header.Get("X-MyHeader") == "test"
		w.Write([]byte(`{"ping": "pong"}`))
	}))

	defer downstreamServer.Close()

	pingRoute := routes.Route{
		IncomingRequestPath:  "/test/ping",
		ForwardedRequestURL:  downstreamServer.URL,
		ForwardedRequestPath: "/ping",
	}

	routesConfig := &routes.RoutesConfig{
		Routes: []routes.Route{pingRoute},
	}

	proxy := NewProxy(logger, routesConfig, 8080)
	proxy.RequestMiddleware = append(proxy.RequestMiddleware, middleware.RequestMiddleware(addMyHeader))

	go func() {
		proxy.Start()
	}()
	defer proxy.Stop()

	resp, err := http.Get("http://localhost:8080/test/ping")
	assert.Nil(err)

	body, err := ioutil.ReadAll(resp.Body)
	assert.Nil(err)

	assert.Equal(`{"ping": "pong"}`, string(body))
	assert.True(headerPresent)
}

func TestProxyRoundtripMiddleware(t *testing.T) {
	assert := assert.New(t)
	logger := &logger.NullLogger{}

	pingRoute := routes.Route{
		IncomingRequestPath:  "/test/ping",
		ForwardedRequestURL:  "http://localhost:3000",
		ForwardedRequestPath: "/ping",
	}

	routesConfig := &routes.RoutesConfig{
		Routes: []routes.Route{pingRoute},
	}

	proxy := NewProxy(logger, routesConfig, 8080)

	requestCounter := 0

	middlewareFunc := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCounter++
			next.ServeHTTP(w, r)
		})
	}

	mw := middleware.RoundtripMiddleware(middlewareFunc)
	proxy.RoundtripMiddleware = append(proxy.RoundtripMiddleware, mw)

	go func() {
		proxy.Start()
	}()
	defer proxy.Stop()

	_, err := http.Get("http://localhost:8080/test/ping")
	assert.Nil(err)
	assert.Equal(1, requestCounter)
}
