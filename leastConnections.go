package main

import (
	"cmp"
	"fmt"
	"net/http"
	"slices"
)

type LeastConnectionsBalancer struct {
	backends []*BackendServer
	client   http.Client
}

func NewLeastConnectionsBalancer() Balancer {
	return &BasicBalancer{
		backends: []*BackendServer{},
		client:   http.Client{},
	}
}

func (lc LeastConnectionsBalancer) ListServers() []*BackendServer {
	return lc.backends
}

func (lc *LeastConnectionsBalancer) NextServer() *BackendServer {
	healthyServers := Filter(lc.backends, func(server *BackendServer) bool {
		return server.IsHealthy
	})
	if len(healthyServers) == 0 {
		return nil
	}
	return slices.MinFunc(lc.backends, func(backend1, backend2 *BackendServer) int {
		return cmp.Compare(backend1.connections.Load(), backend2.connections.Load())
	})
}

func (lc *LeastConnectionsBalancer) Serve(server *BackendServer, req *http.Request) (*http.Response, error) {
	return lc.client.Do(updateRequest(req, *server.HealthCheckEndpoint))
}

func (lc *LeastConnectionsBalancer) RegisterServers(newServers ...*BackendServer) error {

	for _, newServer := range newServers {
		for _, serverLatency := range lc.backends {
			if serverLatency.HealthCheckEndpoint.String() == newServer.HealthCheckEndpoint.String() {
				return fmt.Errorf("provided server has already been registered (healthpoint taken)")
			}
		}
		lc.backends = append(lc.backends, newServer)
	}
	return nil
}

func (lc *LeastConnectionsBalancer) DeregisterServer(removeServer *BackendServer) error {
	for i, serverLatency := range lc.backends {
		if serverLatency.HealthCheckEndpoint.Path == removeServer.HealthCheckEndpoint.Path {
			lc.backends[i] = lc.backends[len(lc.backends)-1]
			lc.backends = lc.backends[:len(lc.backends)-1]
		}
	}
	return fmt.Errorf("Provided server could not be de-registered. It may have already been deregistered.")
}
