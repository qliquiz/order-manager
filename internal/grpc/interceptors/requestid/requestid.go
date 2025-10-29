package requestid

import (
	"context"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const HeaderRequestID = "x-request-id"

func Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		var requestID string

		if ok {
			values := md.Get(HeaderRequestID)
			if len(values) > 0 {
				requestID = values[0]
			}
		}
		if requestID == "" {
			requestID = uuid.New().String()
		}

		ctx = context.WithValue(ctx, HeaderRequestID, requestID)

		err := grpc.SetHeader(ctx, metadata.Pairs(HeaderRequestID, requestID))
		if err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}
