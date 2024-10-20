package main

import (
	"fmt"
	"io"
	Balancers "load-balancer/balancers"
	Monitor "load-balancer/monitor"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

type Config struct {
	LBPort         int                //
	Type           Balancers.Strategy //
	MaxConnections int                // max connections to handle
	MaxBodySize    int                // max body size allowed in bytes
	MonitorConfig  Monitor.Config
}

func main() {
	config := Config{
		LBPort: 8080,
		Type:   Balancers.StrategyRoundRobin,
		MonitorConfig: Monitor.Config{
			MaxAttempts: 1,
			Timeout:     60,
			Protocol:    Monitor.ProtocolHTTP2,
		},
	}

	balancer := Balancers.NewBalancer(config.Type)

	url, err := url.Parse("http://localhost:8000")
	handleErr(err)

	url_2, err := url.Parse("http://localhost:8001")
	handleErr(err)

	server := Balancers.BackendServer{IsHealthy: true, HealthCheckEndpoint: url}
	server2 := Balancers.BackendServer{IsHealthy: true, HealthCheckEndpoint: url_2}
	err = balancer.RegisterServers(&server, &server2)
	handleErr(err)

	m := Monitor.NewMonitor(balancer, config.MonitorConfig)

	go func() {
		for {
			m.CheckHealth()
			time.Sleep(1 * time.Second)
		}
	}()

	http.HandleFunc("/", handleProxy(balancer))
	log.Fatal(http.ListenAndServe(createAddr(config.LBPort), nil))
}

func createAddr(port int) string {
	return fmt.Sprintf(":%d", port)
}

func handleProxy(b Balancers.Balancer) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		slog.Info("incoming request", slog.String("host", r.Host), slog.String("method", r.Method), slog.String("remoteAddr", r.RemoteAddr))

		res, err := b.Serve(r)

		processingTime := time.Since(startTime).String()
		w.Header().Set("X-Processing-Time", processingTime)

		if err != nil {
			slog.Info("not successful", slog.String("err", err.Error()))
			w.WriteHeader(502)
			w.Write([]byte(err.Error()))
		} else {
			defer res.Body.Close()
			slog.Info("what did we get by firing call", slog.Int("statusCode", res.StatusCode))
			w.Header().Set("Content-Type", res.Header.Get("Content-Type"))
			w.WriteHeader(res.StatusCode)
			io.Copy(w, res.Body)
		}
	}
}

func handleErr(err error) {
	if err != nil {
		panic(err)
	}
}
