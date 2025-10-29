package app

import (
	grpcApp "349877-artemkagor05-course-1478/internal/app/grpc"
	orderStore "349877-artemkagor05-course-1478/internal/storage/order"
	"log/slog"
)

type App struct {
	GrpcApp *grpcApp.App
}

func New(log *slog.Logger, grpcPort int) *App {
	storage := orderStore.New()

	GrpcApp := grpcApp.New(storage, log, grpcPort)

	return &App{GrpcApp}
}
