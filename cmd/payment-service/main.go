package main

import (
	"os"

	paymentapp "go-micro/internal/app/payment"
)

func main() {
	os.Exit(run(paymentapp.Run))
}

func run(exec func() error) int {
	if err := exec(); err != nil {
		return 1
	}
	return 0
}
