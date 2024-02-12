package server

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/equinix-labs/otel-init-go/otelinit"
	"go.uber.org/zap"

	rootCmd "github.com/metal-toolbox/fleet-rest-skeleton/cmd"
	"github.com/metal-toolbox/fleet-rest-skeleton/internal/app"
	"github.com/metal-toolbox/fleet-rest-skeleton/internal/metrics"
	"github.com/metal-toolbox/fleet-rest-skeleton/internal/version"
	"github.com/metal-toolbox/fleet-rest-skeleton/pkg/api/routes"
	"github.com/spf13/cobra"
)

var shutdownTimeout = 10 * time.Second

// install server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run API service",
	Run: func(c *cobra.Command, args []string) {
		cfg, err := app.LoadConfiguration(rootCmd.CfgFile)
		if err != nil {
			log.Fatalf("loading configuration: %s", err.Error())
		}

		logger := app.GetLogger(cfg.DeveloperMode)
		//nolint:errcheck
		defer logger.Sync()

		// XXX: Read NATS and or FleetDB Config

		// XXX: add NATS client
		// XXX: add FleetDB client

		ctx, appCancel := context.WithCancel(c.Context())
		app := app.NewApp(ctx, cfg, logger)

		metrics.ListenAndServe()

		// the ignored parameter here is a context annotated with otel-init-go configuration
		_, otelShutdown := otelinit.InitOpenTelemetry(c.Context(), "skeleton-api-server")

		logger.Info("app initialized",
			zap.String("version", version.Current().String()),
		)

		srv := routes.ComposeHTTPServer(app)
		go func() {
			if err := srv.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
				logger.Fatal("error serving API",
					zap.Error(err),
				)
			}
		}()

		app.WaitForSignal()
		logger.Info("signaled to terminate")
		appCancel()

		// call server shutdown with timeout
		ctx, cancel := context.WithTimeout(c.Context(), shutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			logger.Fatal("server shutdown error",
				zap.Error(err),
			)
		}
		otelShutdown(ctx)
		logger.Info("OK, done.")
	},
}

// install command flags
func init() {
	rootCmd.RootCmd.AddCommand(serverCmd)
}
