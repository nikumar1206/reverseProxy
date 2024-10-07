package balancers

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
	// ... more can be added
)

type StrategyBalancer struct {
	strategy Strategy
}

type Balancer interface {
	GetName() string                             // tells u the
	GetStrategy() Strategy                       // tells u the strategy being used, very un-necessary in a 1-1 mapping, likely remove.
	Serve(*http.Request) (*http.Response, error) // fire the API call to a server
	NextServer() *BackendServer                  // should tell u the next server to call
	RegisterServers(...BackendServer) error      // add backends
	DeRegisterServer(BackendServer) error        // remove backend
}

// each strategy should map to a different balancer

type BackendServer struct {
	IsHealthy           bool
	HealthCheckEndpoint *url.URL
}

func NewBalancer(strat Strategy) Balancer {
	switch strat {
	case StrategyBasic:
		return BasicBalancer{
			backends: &[]BackendServer{},
			client:   http.Client{},
		}
	default:
		panic("lol")
	}
}
