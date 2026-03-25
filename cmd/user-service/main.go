package main

import (
	"os"

	userapp "go-micro/internal/app/user"
)

func main() {
	os.Exit(run(userapp.Run))
}

func run(exec func() error) int {
	if err := exec(); err != nil {
		return 1
	}
	return 0
}
