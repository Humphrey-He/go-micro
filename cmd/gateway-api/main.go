package main

import (
	"os"

	gatewayapp "go-micro/internal/app/gateway"
)

// @title Go-Micro Gateway API
// @version 1.0
// @description ???? API ??
// @host localhost:8080
// @BasePath /
func main() {
	os.Exit(run(gatewayapp.Run))
}

func run(exec func() error) int {
	if err := exec(); err != nil {
		return 1
	}
	return 0
}
