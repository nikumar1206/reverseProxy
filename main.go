package main

import (
	"fmt"
	Balancers "load-balancer/balancers"
	Monitor "load-balancer/monitor"
	"log"
	"net/http"
	"net/url"
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
	fmt.Println("config ", config)

	balancer := Balancers.NewBalancer(config.Type)

	url, err := url.Parse("localhost:8000")
	if err != nil {
		panic(err)
	}
	server := Balancers.BackendServer{IsHealthy: true, HealthCheckEndpoint: url}
	balancer.RegisterServers(server)

	http.HandleFunc("/", handleProxy(balancer))
	log.Fatal(http.ListenAndServe(createAddr(config.LBPort), nil))
}

func createAddr(port int) string {
	return fmt.Sprintf(":%d", port)
}

func handleProxy(b Balancers.Balancer) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		res, err := b.Serve(r)
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println("incoming request", r.Host, r.Method, r.RemoteAddr, res.StatusCode)
	}
}
