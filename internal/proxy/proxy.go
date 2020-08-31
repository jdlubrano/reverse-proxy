package proxy

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/jdlubrano/reverse-proxy/internal/logger"
	"github.com/jdlubrano/reverse-proxy/internal/routes"
)

type Proxy struct {
	RequestMiddleware   []RequestMiddleware
	RoundtripMiddleware []RoundtripMiddleware

	client       *http.Client
	logger       logger.Logger
	routesConfig *routes.RoutesConfig
	port         int
	server       *http.Server
}

func NewProxy(logger logger.Logger, routesConfig *routes.RoutesConfig, port int) *Proxy {
	return &Proxy{
		RequestMiddleware:   []RequestMiddleware{CopyHeaders, CopyBody, CopyContentLength},
		RoundtripMiddleware: []RoundtripMiddleware{},

		client:       &http.Client{},
		logger:       logger,
		routesConfig: routesConfig,
		port:         port,
		server:       nil,
	}
}

func (p *Proxy) Start() {
	proxyMux := http.NewServeMux()

	for _, route := range p.routesConfig.Routes {
		proxyMux.Handle(route.IncomingRequestPath, p.makeHandler(route))
	}

	addr := fmt.Sprintf(":%d", p.port)

	p.server = &http.Server{
		Addr:         addr,
		Handler:      proxyMux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	p.logger.Error(p.server.ListenAndServe().Error())
}

func (p *Proxy) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	return p.server.Shutdown(ctx)
}

func (p *Proxy) makeHandler(route routes.Route) http.Handler {
	handler := p.proxyRouter(route)

	for _, middleware := range p.RoundtripMiddleware {
		handler = middleware(handler)
	}

	return handler
}

func (p *Proxy) prepareRequest(incomingRequest *http.Request, outgoingRequest *http.Request) error {
	middlewareChain := FinishRequestPrep

	for _, middleware := range p.RequestMiddleware {
		middlewareChain = middleware(middlewareChain)
	}

	if err := middlewareChain(incomingRequest, outgoingRequest); err != nil {
		return err
	}

	return nil
}

func (p *Proxy) proxyRouter(route routes.Route) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p.logger.Infof("Handling request: %+v", route)

		query := r.URL.Query()
		outgoingURL := fmt.Sprintf("%s%s?%s", route.ForwardedRequestURL, route.ForwardedRequestPath, query.Encode())

		p.logger.Infof("Making new request %s %s", r.Method, outgoingURL)
		outgoingRequest, err := http.NewRequest(r.Method, outgoingURL, nil)

		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		p.logger.Infof("Preparing request %s %s", outgoingRequest.Method, outgoingURL)
		err = p.prepareRequest(r, outgoingRequest)

		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		p.logger.Infof("Executing request %s %s", outgoingRequest.Method, outgoingRequest.URL)
		resp, err := p.client.Do(outgoingRequest)

		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		for key, values := range resp.Header {
			for _, v := range values {
				w.Header().Add(key, v)
			}
		}

		w.WriteHeader(resp.StatusCode)
		_, err = w.Write(body)

		if err != nil {
			p.logger.Infof("Unexpected error for %s %s: %s", outgoingRequest.Method, outgoingURL, err.Error())
			// What does this mean? What happened to end up in this block?
		}
	})
}
