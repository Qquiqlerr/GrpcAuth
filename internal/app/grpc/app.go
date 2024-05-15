package grpcapp

import (
	authgrpc "GrpcAuth/internal/grpc/auth"
	"fmt"
	"google.golang.org/grpc"
	"log/slog"
	"net"
	"strconv"
)

type App struct {
	log        *slog.Logger
	gRpcServer *grpc.Server
	port       int
}

func New(log *slog.Logger, port int) *App {
	grpcServer := grpc.NewServer()
	authgrpc.Register(grpcServer)
	return &App{
		log:        log,
		gRpcServer: grpcServer,
		port:       port,
	}
}

func (a *App) MustStart() {
	if err := a.Start(); err != nil {
		panic(err)
	}
}

func (a *App) Start() error {
	const op = "grpcapp.Start"
	log := a.log.With(
		slog.String("op", op),
		slog.String("port", strconv.Itoa(a.port)))

	lis, err := net.Listen("tcp", ":"+strconv.Itoa(a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	log.Info("grpc server listening on " + lis.Addr().String())

	if err := a.gRpcServer.Serve(lis); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (a *App) Stop() {
	const op = "grpcapp.Stop"
	log := a.log.With(
		slog.String("op", op))
	log.Info("grpc server stopping")
	a.gRpcServer.GracefulStop()
}
