// Package main implements a load balancer with various load-balancing strategies.
package main

import (
	"net/http"
	"net/url"
	"sync/atomic"
)

// Strategy represents different load balancing strategies.
type Strategy int

const (
	// StrategyBasic routes all requests to the first healthy server. This strategy
	// typically results in all requests being sent to the same server.
	StrategyBasic Strategy = iota

	// StrategyRoundRobin distributes requests across available servers in a
	// round-robin fashion. Each server is called in order, looping back to the first.
	StrategyRoundRobin

	// StrategyWeightedRoundRobin distributes requests based on server weights. Servers
	// with higher weights receive more requests.
	StrategyWeightedRoundRobin

	// StrategyLeastConnections sends requests to the server with the fewest active connections.
	StrategyLeastConnections

	// StrategyRandom picks a server at random for each request.
	StrategyRandom

	// StrategyLeastLatency routes requests to the server with the lowest latency.
	StrategyLeastLatency

	// Additional strategies can be added as needed.
)

// BalancerMetadata stores metadata related to a load balancer strategy.
type BalancerMetadata struct {
	BalancerName    string          // Name of the balancer
	StrategyName    string          // Name of the strategy
	NewBalancerFunc func() Balancer // Constructor function for creating a new balancer
}

// StrategyBalancerMap maps each Strategy to its corresponding BalancerMetadata.
var StrategyBalancerMap = map[Strategy]BalancerMetadata{
	StrategyBasic: {
		BalancerName:    "Basic Balancer",
		StrategyName:    "Basic Strategy",
		NewBalancerFunc: NewBasicBalancer,
	},
	StrategyRoundRobin: {
		BalancerName:    "RoundRobin Balancer",
		StrategyName:    "RoundRobin Strategy",
		NewBalancerFunc: NewRoundRobinBalancer,
	},
	StrategyLeastLatency: {
		BalancerName:    "LeastLatency Balancer",
		StrategyName:    "LeastLatency Strategy",
		NewBalancerFunc: NewLeastLatencyBalancer,
	},
	StrategyLeastConnections: {
		BalancerName:    "LeastConnections Balancer",
		StrategyName:    "LeastConnections Strategy",
		NewBalancerFunc: NewLeastLatencyBalancer,
	},
}

// Balancer defines an interface for a load balancer with operations to manage servers
// and route requests.
type Balancer interface {
	// ListServers returns a slice of all registered backend servers.
	ListServers() []*BackendServer

	// NextServer selects and returns the next server for handling a request,
	// based on the balancer's strategy.
	NextServer() *BackendServer

	// Serve sends an HTTP request to the specified BackendServer and returns the
	// response or an error if the request fails.
	Serve(server *BackendServer, req *http.Request) (*http.Response, error)

	// RegisterServers adds one or more BackendServers to the balancer's pool.
	RegisterServers(...*BackendServer) error

	// DeregisterServer removes a specific BackendServer from the balancer's pool.
	DeregisterServer(server *BackendServer) error
}

// BackendServer represents a backend server in the load balancer's pool.
type BackendServer struct {
	// IsHealthy indicates if the server is healthy and can handle requests
	IsHealthy bool

	// HealthCheckEndpoint is the URL endpoint for health checks
	HealthCheckEndpoint *url.URL

	// latency of the server in milliseconds (used by certain strategies)
	latency int64

	// connections maintained by the server (used by certain strategies)
	connections atomic.Uint32
}

func NewBalancer(s Strategy) Balancer {
	return StrategyBalancerMap[s].NewBalancerFunc()
}
