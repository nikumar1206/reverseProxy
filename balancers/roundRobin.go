package balancers

import (
	"fmt"
	"net/http"
)

type RoundRobinBalancer struct {
	backends []*BackendServer
	client   http.Client
	runIndex int
}

func (rr RoundRobinBalancer) GetName() string {
	return "RoundRobin Balancer"
}

func (rr RoundRobinBalancer) GetStrategy() Strategy {
	return StrategyRoundRobin
}

func (rr RoundRobinBalancer) ListServers() []*BackendServer {
	return rr.backends
}

func (rr *RoundRobinBalancer) incrementCounter() {
	rr.runIndex++
	if rr.runIndex >= len(rr.backends) {
		rr.runIndex = 0
	}
}

func (rr *RoundRobinBalancer) NextServer() *BackendServer {
	rr.incrementCounter()
	nextServer := rr.backends[rr.runIndex]
	var counter int
	for !nextServer.IsHealthy && counter < len(rr.backends) {
		rr.incrementCounter()
		nextServer = rr.backends[rr.runIndex]
		counter++
	}
	if nextServer.IsHealthy {
		return nextServer
	}

	return nil
}

func (rr *RoundRobinBalancer) Serve(req *http.Request) (*http.Response, error) {
	server := rr.NextServer()
	if server == nil {
		return nil, fmt.Errorf("no healthy upstream")
	}
	fmt.Println("firing a request against ", server.HealthCheckEndpoint.String())
	return rr.client.Do(updateRequest(req, *server.HealthCheckEndpoint))
}

func (rr *RoundRobinBalancer) RegisterServers(newServers ...*BackendServer) error {

	for _, newServer := range newServers {
		for _, server := range rr.backends {
			if server.HealthCheckEndpoint.String() == newServer.HealthCheckEndpoint.String() {
				return fmt.Errorf("provided server has already been registered (healthpoint taken)")
			}
		}
		rr.backends = append(rr.backends, newServer)
	}
	return nil
}

func (rr RoundRobinBalancer) DeregisterServer(removeServer BackendServer) error {
	for i, server := range rr.backends {
		if server.HealthCheckEndpoint.Path == removeServer.HealthCheckEndpoint.Path {
			rr.backends[i] = rr.backends[len(rr.backends)-1]
			rr.backends = rr.backends[:len(rr.backends)-1]
		}
	}
	return fmt.Errorf("Provided server could not be de-registered. It may have already been deregistered.")
}
