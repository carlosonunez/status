package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/carlosonunez/status/internal/config"
	"github.com/carlosonunez/status/internal/engine"
	"github.com/carlosonunez/status/internal/getter"
	"github.com/carlosonunez/status/internal/setter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func newStartCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start the status sync daemon",
		RunE: func(cmd *cobra.Command, _ []string) error {
			path := cfgPath
			if path == "" {
				path = config.DefaultPath()
			}

			cfg, err := config.Load(path)
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			eng, err := engine.New(cfg, getter.DefaultRegistry(), setter.DefaultRegistry())
			if err != nil {
				return fmt.Errorf("build engine: %w", err)
			}

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			logrus.WithField("config", path).Info("status daemon starting")
			if err := eng.Run(ctx); err != nil {
				return err
			}
			logrus.Info("status daemon stopped")
			return nil
		},
	}
}
