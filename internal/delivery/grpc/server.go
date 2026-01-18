package grpc

import (
	"context"
	"log/slog"
	"net"

	usersProtos "github.com/ttrtcixy/users-protos/gen/go/users"
	grpcauthservise "github.com/ttrtcixy/users/internal/delivery/grpc/auth"
	"github.com/ttrtcixy/users/internal/delivery/grpc/contracts"
	"github.com/ttrtcixy/users/internal/delivery/grpc/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Config struct {
	Host    string `env:"GRPC_HOST,required,notEmpty"`
	Port    string `env:"GRPC_PORT,required,notEmpty"`
	Network string `env:"GRPC_NETWORK,required,notEmpty"`
	Addr    string
}

type Server struct {
	log *slog.Logger
	cfg *Config

	srv *grpc.Server
	l   net.Listener

	userAuthService usersProtos.UsersAuthServer
	userService     usersProtos.UsersServer
}

func (s *Server) register(gRPC *grpc.Server) {
	usersProtos.RegisterUsersAuthServer(gRPC, s.userAuthService)
	usersProtos.RegisterUsersServer(gRPC, s.userService)
}

func NewGRPCServer(log *slog.Logger, cfg *Config, usecase contracts.UseCase) *Server {
	cfg.Addr = net.JoinHostPort(cfg.Host, cfg.Port)

	return &Server{
		log:             log,
		cfg:             cfg,
		userAuthService: grpcauthservise.NewUserAuthService(context.Background(), log, usecase),
	}
}

func (s *Server) Start(ctx context.Context) (err error) {
	s.log.LogAttrs(nil, slog.LevelInfo, "Starting grpc server on", slog.String("addr", s.cfg.Addr))

	s.srv = grpc.NewServer(
		grpc.UnaryInterceptor(middleware.RecoveryUnaryInterceptor(s.log)),
	)

	s.register(s.srv)
	s.l, err = net.Listen(s.cfg.Network, s.cfg.Addr)
	if err != nil {
		return err
	}
	reflection.Register(s.srv)

	return s.srv.Serve(s.l)
}

func (s *Server) Close(ctx context.Context) error {
	const op = "grpc.Close()"

	s.srv.GracefulStop()

	return nil
}
