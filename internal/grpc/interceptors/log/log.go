package log

import (
	"349877-artemkagor05-course-1478/internal/grpc/interceptors/requestid"
	"context"
	"google.golang.org/grpc"
	"log/slog"
	"time"
)

func Unary(log *slog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		reqID, _ := ctx.Value(requestid.HeaderRequestIDKey).(string)

		reqLog := log.With(
			slog.String("method", info.FullMethod),
			slog.String("request_id", reqID),
		)

		startTime := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(startTime)

		if err != nil {
			reqLog.Error("request failed",
				slog.Any("request", req),
				slog.String("error", err.Error()),
				slog.Duration("duration", duration),
			)
		} else {
			reqLog.Debug("request completed",
				slog.Any("request", req),
				slog.Any("response", resp),
				slog.Duration("duration", duration),
			)
		}

		return resp, err
	}
}
