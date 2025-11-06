package app

import (
	"349877-artemkagor05-course-1478/internal/app/gateway"
	grpcApp "349877-artemkagor05-course-1478/internal/app/grpc"
	orderrepo "349877-artemkagor05-course-1478/internal/repository/order"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"time"
)

type App struct {
	GrpcApp    *grpcApp.App
	GatewayApp *gateway.App
}

func New(db *pgxpool.Pool, log *slog.Logger, grpcPort int, gatewayPort int, timeout time.Duration) *App {
	repo := orderrepo.New(db)

	grpcApplication := grpcApp.New(repo, log, grpcPort, timeout)
	gatewayApplication := gateway.New(log, gatewayPort, grpcPort)

	return &App{
		GrpcApp:    grpcApplication,
		GatewayApp: gatewayApplication,
	}
}
