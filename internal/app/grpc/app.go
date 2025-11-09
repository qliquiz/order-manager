package grpcapp

import (
	"349877-artemkagor05-course-1478/internal/cache"
	interceptorLog "349877-artemkagor05-course-1478/internal/grpc/interceptors/log"
	interceptorRequestID "349877-artemkagor05-course-1478/internal/grpc/interceptors/requestid"
	orderrepo "349877-artemkagor05-course-1478/internal/repository/order"
	orderService "349877-artemkagor05-course-1478/internal/service/order"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log/slog"
	"net"
	"time"
)

type App struct {
	gRPCServer *grpc.Server
	log        *slog.Logger
	port       int
}

func New(
	repo *orderrepo.OrderRepository,
	cache *cache.OrderCache,
	log *slog.Logger,
	port int,
	timeout time.Duration,
) *App {
	reqIDInterceptor := interceptorRequestID.Unary()
	logInterceptor := interceptorLog.Unary(log)

	gRPCServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			reqIDInterceptor,
			logInterceptor,
		),
	)

	orderService.Register(gRPCServer, repo, cache, log, timeout)

	return &App{
		gRPCServer,
		log,
		port,
	}
}

func (a *App) MustRun() {
	if err := a.run(); err != nil {
		panic(err)
	}
}

func (a *App) run() error {
	const op = "grpcapp.run"
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	a.log.With(slog.String("op", op)).
		Info("grpc server is running %s", "addr", lis.Addr().String())

	if err = a.gRPCServer.Serve(lis); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

func (a *App) Stop(ctx context.Context) error {
	const op = "grpcapp.Stop"
	a.log.With(slog.String("op", op)).
		Info("stopping gRPC server")

	doneChan := make(chan struct{})

	go func() {
		a.gRPCServer.GracefulStop()
		close(doneChan)
	}()

	select {
	case <-doneChan:
		a.log.Info("gRPC server stopped gracefully")
		return nil
	case <-ctx.Done():
		a.log.Warn("gRPC server graceful stop timed out, forcing stop")
		a.gRPCServer.Stop()
		return fmt.Errorf("%s: %w", op, ctx.Err())
	}
}
