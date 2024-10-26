package main

import (
	"cmp"
	"fmt"
	"net/http"
	"slices"
)

type LeastLatencyBalancer struct {
	backends []*BackendServer
	client   http.Client
}

func NewLeastLatencyBalancer() Balancer {
	return &BasicBalancer{
		backends: []*BackendServer{},
		client:   http.Client{},
	}
}

func (ll LeastLatencyBalancer) ListServers() []*BackendServer {
	return ll.backends
}

func (ll *LeastLatencyBalancer) NextServer() *BackendServer {
	healthyBackends := Filter(ll.backends, func(a *BackendServer) bool {
		return a.IsHealthy
	})
	if len(healthyBackends) == 0 {
		return nil
	}
	if len(healthyBackends) == 1 { // should be faster to short circuit here?
		return healthyBackends[0]
	}
	fastestBackend := slices.MinFunc(healthyBackends, func(serverLatency, serverLatency2 *BackendServer) int {
		return cmp.Compare(serverLatency.latency, serverLatency.latency)

	})
	return fastestBackend
}

func (ll *LeastLatencyBalancer) Serve(server *BackendServer, req *http.Request) (*http.Response, error) {
	fmt.Println("firing a request against ", server.HealthCheckEndpoint.String())
	return ll.client.Do(updateRequest(req, *server.HealthCheckEndpoint))
}

func (ll *LeastLatencyBalancer) RegisterServers(newServers ...*BackendServer) error {

	for _, newServer := range newServers {
		for _, serverLatency := range ll.backends {
			if serverLatency.HealthCheckEndpoint.String() == newServer.HealthCheckEndpoint.String() {
				return fmt.Errorf("provided server has already been registered (healthpoint taken)")
			}
		}
		ll.backends = append(ll.backends, newServer)
	}
	return nil
}

func (ll *LeastLatencyBalancer) DeregisterServer(removeServer *BackendServer) error {
	for i, serverLatency := range ll.backends {
		if serverLatency.HealthCheckEndpoint.Path == removeServer.HealthCheckEndpoint.Path {
			ll.backends[i] = ll.backends[len(ll.backends)-1]
			ll.backends = ll.backends[:len(ll.backends)-1]
		}
	}
	return fmt.Errorf("Provided server could not be de-registered. It may have already been deregistered.")
}
