package app

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"

	"github.com/ttrtcixy/users/internal/app/provider"
)

// App represents the main application container.
type App struct {
	// workServerCount tracks the number of currently running goroutines/servers.
	workServerCount atomic.Int64
	*provider.Provider
	// stop is a buffered channel used to capture OS termination signals (SIGINT, SIGTERM).
	stop chan os.Signal

	// serverErr receives errors from individual server goroutines.
	serverErr chan serverError

	// app close signal.
	shutdown context.CancelFunc
}

// New initializes the application, sets up signal notification, and returns an App instance.
func New(ctx context.Context) *App {
	const op = "app.New()"

	ctx, cancel := context.WithCancel(ctx)

	// Initialize the provider (database, logger, config, etc.).
	p, err := provider.New(ctx)
	if err != nil {
		// Using Fatalf because the app cannot function without a valid provider.
		log.Fatalf("%s - provider initialization failed -> %s", op, err.Error())
	}

	// Channel to listen for system interruptions.
	closeChan := make(chan os.Signal, 1)
	signal.Notify(closeChan, os.Interrupt, syscall.SIGTERM)

	return &App{
		Provider:  p,
		stop:      closeChan,
		serverErr: make(chan serverError, 1),
		shutdown:  cancel,
	}
}

// serverError wraps a standard error message with metadata about which server failed.
type serverError struct {
	message string
	serverInfo
}

// Error implements the error interface.
func (e serverError) Error() string {
	return "server error"
}

// serverInfo contains metadata used to determine how to handle a server crash.
type serverInfo struct {
	serverName string
	// If true, a crash in this server forces the whole app to exit
	errStopApp bool
}

// Run starts the application's main event loop.
func (a *App) Run(ctx context.Context) {

	a.workServerCount.Add(1)
	go a.runServer(ctx, serverInfo{serverName: "http_server", errStopApp: false}, a.GRPCServer.Start)

	// Listen signals or errors
	for {
		select {
		case <-a.stop:
			a.shutdown()
			a.Closer.Close()
			os.Exit(0)
		// Scenario 2: One of the servers crashed or stopped with an error.
		case err := <-a.serverErr:
			a.Logger.LogAttrs(nil, slog.LevelError,
				"Server crashed",
				slog.String("server", err.serverName),
				slog.String("error", err.message),
				slog.Bool("close_app", err.errStopApp))

			// If the specific server is critical (errStopApp) or no servers are left running, exit.
			if err.errStopApp || a.workServerCount.Load() == 0 {
				a.Closer.Close()
				os.Exit(0)
			}
		}
	}
}

// runServer is a helper that wraps a server's Start function with error handling and tracking.
func (a *App) runServer(ctx context.Context, serverInfo serverInfo, fn func(ctx context.Context) error) {
	if err := fn(ctx); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			a.workServerCount.Add(-1)
			return
		}

		sErr := serverError{
			message:    err.Error(),
			serverInfo: serverInfo,
		}
		a.workServerCount.Add(-1)
		a.serverErr <- sErr
	}
}
