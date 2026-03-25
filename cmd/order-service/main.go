package main

import (
	"os"

	orderapp "go-micro/internal/app/order"
)

func main() {
	os.Exit(run(orderapp.Run))
}

func run(exec func() error) int {
	if err := exec(); err != nil {
		return 1
	}
	return 0
}
