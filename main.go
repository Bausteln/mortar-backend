package main

import (
	"log"

	"gitlab.bausteln.ch/net-core/reverse-proxy/mortar-backend/internal/k8s"
	"gitlab.bausteln.ch/net-core/reverse-proxy/mortar-backend/internal/server"
)

func main() {
	// Create Kubernetes dynamic client
	dynamicClient, err := k8s.NewDynamicClient()
	if err != nil {
		log.Fatalf("Error creating Kubernetes client: %v", err)
	}

	// Create and start server
	srv := server.New("8080", dynamicClient)
	srv.Run()
}
