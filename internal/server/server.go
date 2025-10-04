package server

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"gitlab.bausteln.ch/net-core/reverse-proxy/mortar-backend/internal/handlers"
	"k8s.io/client-go/dynamic"
)

type Server struct {
	port              string
	proxyRulesHandler *handlers.ProxyRulesHandler
}

func New(port string, dynamicClient dynamic.Interface) *Server {
	return &Server{
		port:              port,
		proxyRulesHandler: handlers.NewProxyRulesHandler(dynamicClient),
	}
}

func (s *Server) Start() error {
	// Register routes
	http.HandleFunc("/api/proxyrules", s.handleProxyRules)
	http.HandleFunc("/api/proxyrules/", s.handleProxyRules)

	// Start server
	fmt.Printf("Starting API server on port %s...\n", s.port)
	if err := http.ListenAndServe(":"+s.port, nil); err != nil {
		return fmt.Errorf("error starting server: %w", err)
	}
	return nil
}

func (s *Server) handleProxyRules(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")

	// /api/proxyrules -> list all
	if len(parts) == 2 && parts[1] == "proxyrules" {
		s.proxyRulesHandler.GetProxyRules(w, r)
		return
	}

	// /api/proxyrules/{name} -> get specific
	if len(parts) == 3 && parts[1] == "proxyrules" {
		s.proxyRulesHandler.GetProxyRule(w, r)
		return
	}

	http.Error(w, "Not found", http.StatusNotFound)
}

func (s *Server) Run() {
	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
}
