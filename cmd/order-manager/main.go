package main

import (
	"349877-artemkagor05-course-1478/internal/app"
	"349877-artemkagor05-course-1478/internal/config"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := config.MustLoad()

	application := app.New(cfg.GRPC.Port)
	go application.GrpcApp.MustRun()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sign := <-stop

	application.GrpcApp.Stop()
	log.Printf("application stopped with signal: %s", sign.String())
}
