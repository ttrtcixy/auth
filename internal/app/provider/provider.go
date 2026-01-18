package provider

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	logger "github.com/ttrtcixy/color-slog-handler"
	storage "github.com/ttrtcixy/pgx-wrapper"
	closer "github.com/ttrtcixy/task-closer"
	"github.com/ttrtcixy/users/internal/config"
	"github.com/ttrtcixy/users/internal/delivery/grpc"
	domainservice "github.com/ttrtcixy/users/internal/domain/service"
	"github.com/ttrtcixy/users/internal/infrastructure/service"
	"github.com/ttrtcixy/users/internal/repository"
	"github.com/ttrtcixy/users/internal/usecase"
)

type Provider struct {
	Logger *slog.Logger
	Closer closer.Closer

	cfg *config.Config

	db             *storage.Postgres
	infraServices  *infraservice.InfraServices
	domainServices *domainservice.DomainService

	usecase    *usecase.UseCase
	repository *repository.Repository

	GRPCServer *grpc.Server
}

func New(ctx context.Context) (p *Provider, err error) {
	const op = "provider.New()"
	defer func() {
		if err != nil {
			err = fmt.Errorf("%s -> %w", op, err)
		}
	}()

	p = &Provider{}

	if err = p.setupCore(ctx); err != nil {
		return p, err
	}

	if err = p.setupInfrastructure(ctx); err != nil {
		return p, err
	}

	if err = p.setupApplication(ctx); err != nil {
		return p, err
	}

	p.GRPCServer = grpc.NewGRPCServer(p.Logger, &p.cfg.GRPC, p.usecase)

	p.Closer.Add("stop_grpc_server", p.GRPCServer.Close)

	return p, nil
}

func (p *Provider) setupCore(_ context.Context) error {
	const op = "provider.setupCore()"
	// config
	cfg, err := config.New()
	if err != nil {
		return fmt.Errorf("%s - config init failed -> %w", op, err)
	}
	p.cfg = cfg

	// Logger
	//textHandler := logger.NewTextHandler(os.Stdout, &logger.Config{Level: p.cfg.Logger.Level, BufferedOutput: true})
	textHandler := logger.NewTextHandler(os.Stderr, &p.cfg.Logger)
	p.Logger = slog.New(textHandler)

	p.Logger.LogAttrs(nil, slog.LevelInfo, "Logger load", slog.String("level", logger.ParseLevel(p.cfg.Logger.Level)))

	// Closer
	// Create a new Logger for Closer. IMPORTANT: Output may be problematic because the default Logger and Closer Logger might write to the same writer (if one is buffered, output could be fragmented).
	closerTextHandler := logger.NewTextHandler(os.Stdout, &logger.Config{Level: p.cfg.Logger.Level, BufferedOutput: false})
	p.Closer = closer.New(slog.New(closerTextHandler).WithGroup("closer"), &p.cfg.Closer)

	p.Closer.Add(
		"flush_logger",
		textHandler.Close,
	)

	p.Closer.Add(
		"env_clear",
		p.cfg.Close,
	)

	return nil
}

func (p *Provider) setupInfrastructure(ctx context.Context) error {
	const op = "provider.setupInfrastructure()"
	// db
	db, err := storage.New(ctx, p.Logger, &p.cfg.DB)
	if err != nil {
		return fmt.Errorf("%s - db init failed -> %w", op, err)
	}
	p.db = db
	p.Logger.LogAttrs(nil, slog.LevelInfo, "Connect to database successful")

	p.Closer.Add(
		"close_db_connection",
		p.db.Close,
	)

	p.infraServices = infraservice.New(&p.cfg.InfraService)
	p.domainServices = domainservice.NewDomainService(&p.cfg.DomainService, p.repository, p.infraServices)

	return nil
}

func (p *Provider) setupApplication(ctx context.Context) error {
	const op = "provider.setupApplication()"
	// repository
	p.repository = repository.NewRepository(ctx, p.Logger, p.db)

	// usecase
	p.usecase = usecase.NewUseCase(p.Logger, p.repository, p.infraServices, p.domainServices)
	return nil
}
