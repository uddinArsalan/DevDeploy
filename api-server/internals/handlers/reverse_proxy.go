package handlers

import (
	"net/http"
	"net/http/httputil"

	"github.com/uddinArsalan/devdeploy/internals/services"
)

type ProxyHandler struct {
	ps *services.ProxyService
}

func NewProxyHandler(proxyService *services.ProxyService) *ProxyHandler {
	return &ProxyHandler{
		ps: proxyService,
	}
}

func (h *ProxyHandler) ReverseHandler(w http.ResponseWriter, r *http.Request) {
	target, err := h.ps.Route(r.Context(),r.Host)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ServeHTTP(w, r)
}
