package proxy

import (
	"net/http"
)

type RoundtripMiddleware func(next http.Handler) http.Handler
