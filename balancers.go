package main

import (
	"net/http"
	"net/url"
)

type Strategy int

const (
	StrategyBasic              Strategy = iota // StrategyBasic just fires request at the first healthy server. lol, all requests go to same server... usually.
	StrategyRoundRobin                         // TODO
	StrategyWeightedRoundRobin                 // TODO
	StrategyLeastConnections                   // TODO
	StrategyRandom                             // TODO
	StrategyLeastLatency                       // TODO
	// ... more can be added
)

type Balancer interface {
	GetName() string                             // tells u the name of the balancer
	GetStrategy() Strategy                       // tells u the strategy being used, very un-necessary in a 1-1 mapping, likely remove.
	Serve(*http.Request) (*http.Response, error) // fire the API call to a server
	NextServer() *BackendServer                  // should tell u the next server to call
	RegisterServers(...*BackendServer) error     // add backends
	DeregisterServer(BackendServer) error        // remove backend
	ListServers() []*BackendServer
}

// each strategy should map to a different balancer
// might need a unique identifier for each server
type BackendServer struct {
	IsHealthy           bool
	HealthCheckEndpoint *url.URL
	latency             int64
	connections         uint8
}

func NewBalancer(strat Strategy) Balancer {
	switch strat {
	case StrategyBasic:
		return &BasicBalancer{
			backends: []*BackendServer{},
			client:   http.Client{},
		}
	case StrategyRoundRobin:
		return &RoundRobinBalancer{
			backends: []*BackendServer{},
			client:   http.Client{},
		}
	case StrategyLeastLatency:
		return &LeastLatencyBalancer{
			backends: []*BackendServer{},
			client:   http.Client{},
		}
	default:
		panic("lol")
	}
}
