package main

import (
	"os"

	inventoryapp "go-micro/internal/app/inventory"
)

func main() {
	os.Exit(run(inventoryapp.Run))
}

func run(exec func() error) int {
	if err := exec(); err != nil {
		return 1
	}
	return 0
}
