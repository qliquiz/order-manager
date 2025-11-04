package gateway

import (
	"349877-artemkagor05-course-1478/gen/api"
	"context"
	"errors"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log/slog"
	"net/http"
)

type App struct {
	server     *http.Server
	log        *slog.Logger
	port       int
	grpcTarget string
}

func New(log *slog.Logger, port int, grpcPort int) *App {
	grpcTarget := fmt.Sprintf("localhost:%d", grpcPort)

	return &App{
		log:        log,
		port:       port,
		grpcTarget: grpcTarget,
	}
}

func (a *App) MustRun() {
	if err := a.run(); err != nil {
		panic(err)
	}
}

func (a *App) run() error {
	const op = "gateway.App.run"
	log := a.log.With(slog.String("op", op), slog.Int("port", a.port))

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := api.RegisterOrderServiceHandlerFromEndpoint(ctx, mux, a.grpcTarget, opts)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("starting gateway")

	a.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", a.port),
		Handler: mux,
	}

	if err = a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) Stop(ctx context.Context) error {
	const op = "gateway.App.Stop"
	log := a.log.With(slog.String("op", op))

	log.Info("stopping gateway")

	return a.server.Shutdown(ctx)
}
