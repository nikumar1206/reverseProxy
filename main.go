package main

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

type Config struct {
	LBPort         int      //
	Type           Strategy //
	MaxConnections int      // max connections to handle
	MaxBodySize    int      // max body size allowed in bytes
	MonitorConfig  MonitorConfig
}

func main() {
	config := Config{
		LBPort: 8080,
		Type:   StrategyLeastConnections,
		MonitorConfig: MonitorConfig{
			MaxAttempts: 1,
			Timeout:     60,
			Protocol:    ProtocolHTTP11,
		},
	}

	balancer := NewBalancer(config.Type)

	url, err := url.Parse("http://localhost:8000")
	handleErr(err)

	url_2, err := url.Parse("http://localhost:8001")
	handleErr(err)
	url_3, err := url.Parse("http://localhost:8003")
	handleErr(err)

	server := BackendServer{IsHealthy: true, HealthCheckEndpoint: url}
	server2 := BackendServer{IsHealthy: true, HealthCheckEndpoint: url_2}
	server3 := BackendServer{IsHealthy: true, HealthCheckEndpoint: url_3}
	err = balancer.RegisterServers(&server, &server2, &server3)
	handleErr(err)

	m := NewMonitor(balancer, config.MonitorConfig)

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

func handleProxy(b Balancer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		slog.Info("incoming request", slog.String("host", r.Host), slog.String("method", r.Method), slog.String("remoteAddr", r.RemoteAddr))

		server := b.NextServer()
		if server == nil {
			handleRequestError(fmt.Errorf("no healthy upstream"), w)
			return
		}
		slog.Info(fmt.Sprintf("using server %s", server))
		server.connections.Add(1)
		defer server.connections.Add(^uint32(0))
		res, err := b.Serve(server, r)

		processingTime := time.Since(startTime).String()
		w.Header().Set("X-Processing-Time", processingTime)

		if err != nil {
			handleRequestError(err, w)
			return
		}
		defer res.Body.Close()
		slog.Info("Status Code from initial call", slog.Int("statusCode", res.StatusCode))
		w.Header().Set("Content-Type", res.Header.Get("Content-Type"))
		w.WriteHeader(res.StatusCode)

		buf := make([]byte, 1024)
		for {
			n, err := res.Body.Read(buf)
			if n > 0 {
				if _, writeErr := w.Write(buf[:n]); writeErr != nil {
					slog.Error("Error writing to response", slog.Any("error", writeErr))
					return
				}
				// Flush to ensure the data is sent as soon as it's written
				if flusher, ok := w.(http.Flusher); ok {
					flusher.Flush()
				}
			}
			if err != nil {
				if err == io.EOF {
					break
				}
				slog.Error("Error reading from source body", slog.Any("error", err))
				return
			}
		}

	}
}

func handleErr(err error) {
	if err != nil {
		panic(err)
	}
}

func handleRequestError(err error, w http.ResponseWriter) {
	slog.Info("not successful", slog.String("err", err.Error()))
	w.WriteHeader(502)
	w.Write([]byte(err.Error()))
}
