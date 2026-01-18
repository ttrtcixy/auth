package middleware

import (
	"context"
	"fmt"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// todo refactor
func RecoveryUnaryInterceptor(log *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func(log *slog.Logger) {
			if r := recover(); r != nil {
				e := fmt.Errorf("panic in %s: %v", info.FullMethod, r)
				log.Error(e.Error())
				err = status.Error(codes.Internal, "internal server error")
			}
		}(log)

		return handler(ctx, req)
	}
}
