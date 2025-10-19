package app

import (
	grpcapp "349877-artemkagor05-course-1478/internal/app/grpc"
	orderstore "349877-artemkagor05-course-1478/internal/storage/order"
)

type App struct {
	GrpcApp *grpcapp.App
}

func New(grpcPort int) *App {
	storage := orderstore.New()

	grpcApp := grpcapp.New(storage, grpcPort)

	return &App{
		GrpcApp: grpcApp,
	}
}
