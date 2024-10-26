package main

import (
	"fmt"
	"net/http"
)

type BasicBalancer struct {
	backends []*BackendServer
	client   http.Client
}

func (bb BasicBalancer) GetName() string {
	return "Basic Balancer"
}

func (bb BasicBalancer) GetStrategy() Strategy {
	return StrategyBasic
}

func (bb BasicBalancer) ListServers() []*BackendServer {
	return bb.backends
}

func (bb BasicBalancer) NextServer() *BackendServer {
	for _, server := range bb.backends {
		if server.IsHealthy {
			return server
		}
	}
	return nil
}

func (bb BasicBalancer) Serve(req *http.Request) (*http.Response, error) {
	server := bb.NextServer()
	if server == nil {

		return nil, fmt.Errorf("no healthy upstream")
	}

	return bb.client.Do(updateRequest(req, *server.HealthCheckEndpoint))
}

func (bb *BasicBalancer) RegisterServers(newServers ...*BackendServer) error {

	for _, newServer := range newServers {
		for _, server := range bb.backends {
			if server.HealthCheckEndpoint.String() == newServer.HealthCheckEndpoint.String() {
				return fmt.Errorf("provided server has already been registered (healthpoint taken)")
			}
		}
		bb.backends = append(bb.backends, newServer)
	}
	return nil
}

func (bb BasicBalancer) DeregisterServer(removeServer BackendServer) error {
	for i, server := range bb.backends {
		if server.HealthCheckEndpoint.Path == removeServer.HealthCheckEndpoint.Path {
			bb.backends[i] = bb.backends[len(bb.backends)-1]
			bb.backends = bb.backends[:len(bb.backends)-1]
		}
	}
	return fmt.Errorf("Provided server could not be de-registered. It may have already been deregistered.")
}
