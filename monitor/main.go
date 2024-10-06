package monitor

type Config struct {
	MaxAttempts int    // NumberofAttempts before marking a server as unhealthy
	Timeout     int    // How long to wait for the Server to respond in a health check.
	Protocol    string // http/1.1 or 2 or something else.

}

func checkHealth() {
}
