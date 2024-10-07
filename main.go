package main

import (
	"fmt"
	"io"
	Balancers "load-balancer/balancers"
	Monitor "load-balancer/monitor"
	"log"
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
		Type:   Balancers.StrategyBasic,
	}

	balancer := Balancers.NewBalancer(config.Type)

	url, err := url.Parse("http://localhost:8000")
	if err != nil {
		panic(err)
	}
	server := Balancers.BackendServer{IsHealthy: true, HealthCheckEndpoint: url}
	err = balancer.RegisterServers(server)
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/", handleProxy(balancer))
	log.Fatal(http.ListenAndServe(createAddr(config.LBPort), nil))
}

func createAddr(port int) string {
	return fmt.Sprintf(":%d", port)
}

func handleProxy(b Balancers.Balancer) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		fmt.Println("incoming request", r.Host, r.Method, r.RemoteAddr)

		res, err := b.Serve(r)

		processingTime := time.Since(startTime).String()
		w.Header().Set("X-Processing-Time", processingTime)

		if err != nil {
			fmt.Println("not successful", err.Error())
			w.WriteHeader(502)
			w.Write([]byte(err.Error()))
		} else {
			defer res.Body.Close()
			fmt.Println("what did we get by firing call", res.StatusCode)
			w.Header().Set("Content-Type", res.Header.Get("Content-Type"))
			w.WriteHeader(res.StatusCode)
			io.Copy(w, res.Body)
		}
	}
}
