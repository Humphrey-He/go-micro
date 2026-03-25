package main

import (
	"os"

	taskapp "go-micro/internal/app/task"
)

func main() {
	os.Exit(run(taskapp.Run))
}

func run(exec func() error) int {
	if err := exec(); err != nil {
		return 1
	}
	return 0
}
