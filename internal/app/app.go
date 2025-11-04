package app

import (
	"349877-artemkagor05-course-1478/internal/app/gateway"
	grpcApp "349877-artemkagor05-course-1478/internal/app/grpc"
	orderStore "349877-artemkagor05-course-1478/internal/storage/order"
	"log/slog"
)

type App struct {
	GrpcApp    *grpcApp.App
	GatewayApp *gateway.App
}

func New(log *slog.Logger, grpcPort int, gatewayPort int) *App {
	storage := orderStore.New()

	grpcApplication := grpcApp.New(storage, log, grpcPort)
	gatewayApplication := gateway.New(log, gatewayPort, grpcPort)

	return &App{
		GrpcApp:    grpcApplication,
		GatewayApp: gatewayApplication,
	}
}
