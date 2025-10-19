package grpcapp

import (
	orderservice "349877-artemkagor05-course-1478/internal/service/order"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
)

type App struct {
	gRPCServer *grpc.Server
	port       int
}

func New(order orderservice.Order, port int) *App {
	gRPCServer := grpc.NewServer()

	orderservice.Register(gRPCServer, order)

	return &App{
		gRPCServer,
		port,
	}
}

func (a *App) MustRun() {
	if err := a.run(); err != nil {
		panic(err)
	}
}

func (a *App) run() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	log.Printf("grpc server is running %s", lis.Addr().String())

	if err = a.gRPCServer.Serve(lis); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

func (a *App) Stop() {
	a.gRPCServer.GracefulStop()
}
