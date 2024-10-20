package monitor

import (
	Balancers "load-balancer/balancers"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"golang.org/x/net/http2"
)

type Protocol int

const (
	ProtocolHTTP11 Protocol = iota
	ProtocolHTTP2
)

type Config struct {
	MaxAttempts int      // NumberofAttempts before marking a server as unhealthy
	Timeout     int      // How long to wait for the Server to respond in a health check in seconds. anywhere from 0-300 seconds
	Protocol    Protocol // http/1.1 or 2 or something else.
}

// monitors the backends of the provided balancer config
type Monitor struct {
	balancer Balancers.Balancer
	config   Config
	client   *http.Client
}

func NewMonitor(b Balancers.Balancer, config Config) *Monitor {
	// is it better to validate the config and fix it, or should we throw error.
	var transport http.RoundTripper
	if config.Protocol == ProtocolHTTP2 {
		transport = &http2.Transport{}
	} else {
		transport = &http.Transport{}
	}

	return &Monitor{
		balancer: b,
		config:   config,
		client:   &http.Client{Transport: transport, Timeout: time.Second * time.Duration(config.Timeout)},
	}
}

type ServerHealth struct {
	serverURL string
	isHealthy bool
}

func (m *Monitor) CheckHealth() {
	var wg sync.WaitGroup
	for _, server := range m.balancer.ListServers() {
		wg.Add(1)
		go m.Fire(server, &wg)
	}
	wg.Wait()
}

func (m *Monitor) Fire(server *Balancers.BackendServer, wg *sync.WaitGroup) {
	defer wg.Done()

	serverURL := server.HealthCheckEndpoint.String()
	req, err := http.NewRequest(http.MethodHead, serverURL, nil)
	if err != nil {
		slog.Error("likely misconfiguration in server setup")
		panic(err)
	}

	var totalAttempts int
	var isHealthy bool

	for totalAttempts < m.config.MaxAttempts && !isHealthy {
		isHealthy = m.validateResponse(m.client.Do(req))
		totalAttempts++
	}
	slog.Info("healthcheck", slog.String("serverID", serverURL), slog.Bool("isHealthy", isHealthy))
	server.IsHealthy = isHealthy
}

func (m *Monitor) validateResponse(res *http.Response, err error) bool {
	if err != nil {
		slog.Debug(err.Error())
		return false
	}
	if res.StatusCode > 299 {
		return false
	}
	return true
}
