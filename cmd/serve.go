package cmd

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/trranminhquang/go-boilerplate/internal/api"
	"github.com/trranminhquang/go-boilerplate/internal/conf"
	"golang.org/x/sys/unix"
)

// serveCmd represents the serve command
var serveCmd = cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP server",
	Long:  "Start the server to handle incoming HTTP requests",
	Run: func(cmd *cobra.Command, args []string) {
		startServer(cmd.Context())
	},
}

// startServer initializes and runs the HTTP server
func startServer(ctx context.Context) {
	// Load configuration
	if err := conf.LoadFile(configFile); err != nil {
		logrus.WithError(err).Fatal("Unable to load config file")
	}

	if err := conf.LoadDirectory(watchDir); err != nil {
		logrus.WithError(err).Error("Unable to load config from watch directory")
	}

	config, err := conf.LoadGlobalFromEnv()
	if err != nil {
		logrus.WithError(err).Fatal("Unable to load config from environment")
	}

	// Setup server
	addr := net.JoinHostPort(config.API.Host, config.API.Port)
	apiServer := api.NewApiWithVersion("1.0.0", config)
	logrus.WithField("version", apiServer.Version()).Infof("API starting on: %s", addr)

	// Create base context
	baseCtx, baseCancel := context.WithCancel(context.Background())
	defer baseCancel()

	// Configure HTTP server
	httpServer := &http.Server{
		Addr:              addr,
		Handler:           apiServer,
		ReadHeaderTimeout: 2 * time.Second, // to mitigate Slowloris attacks
		BaseContext: func(net.Listener) context.Context {
			return baseCtx
		},
	}
	logger := logrus.WithField("component", "api-server")

	// Setup graceful shutdown
	var wg sync.WaitGroup
	defer wg.Wait()

	wg.Add(1)
	go handleGracefulShutdown(ctx, &wg, httpServer, baseCancel, logger)

	// Configure socket options and start server
	listener, err := createListener(ctx, addr)
	if err != nil {
		logger.WithError(err).Fatal("HTTP server listen failed")
	}

	if err := httpServer.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.WithError(err).Fatal("HTTP server serve failed")
	}
}

// handleGracefulShutdown performs graceful shutdown when context is done
func handleGracefulShutdown(
	ctx context.Context,
	wg *sync.WaitGroup,
	server *http.Server,
	baseCancel context.CancelFunc,
	logger *logrus.Entry,
) {
	defer wg.Done()

	<-ctx.Done()
	defer baseCancel() // close baseContext

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Minute)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil && !errors.Is(err, context.Canceled) {
		logger.WithError(err).Error("Server shutdown failed")
	}
}

// createListener creates a TCP listener with SO_REUSEPORT option
func createListener(ctx context.Context, addr string) (net.Listener, error) {
	lc := net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) error {
			var socketErr error
			if err := c.Control(func(fd uintptr) {
				socketErr = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
			}); err != nil {
				return err
			}
			return socketErr
		},
	}
	return lc.Listen(ctx, "tcp", addr)
}
