package serve

import (
	"context"
	"fmt"
	"github.com/ohhfishal/fishy/database"
	"github.com/ohhfishal/fishy/notify"
	"log/slog"
	"time"
)

type CMD struct {
	Config ServerConfig `embed:"" group:"Server Config"`
}

type ServerConfig struct {
	Webhook      string              `arg:"" required:"" help:"Discord webhook to send message to."`
	Database     string              `arg:"" default:"fishy.db" help:"SQLite connection string."`
	EmbedOptions notify.EmbedOptions `embed:""`

	CardFile    string        `name:"load" short:"l" type:"existingfile" help:"Generated flashcard file to load in. Ignores duplicates."`
	Interval    time.Duration `default:"15m" help:"Minimum duration between notifications."`
	Heartbeat   time.Duration `default:"1m" help:"Duration between checks if there is work to be done."`
	Probability float64       `default:"0.50" help:"Starting probability a notification is send after interval."`
	Delta       float64       `default:"0.1" help:"Delta added to probability on failure to trigger."`
}

func (cmd *CMD) Run(ctx context.Context, logger *slog.Logger) error {
	return cmd.Config.Run(ctx, logger)
}

func (config *ServerConfig) Run(ctx context.Context, logger *slog.Logger) error {
	db, err := database.Connect(ctx, "sqlite", config.Database)
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}

	if path := config.CardFile; path != "" {
		logger.Warn("implement card loading", "file", path)
	}

	metrics, err := db.Metrics(ctx)
	if err != nil {
		return fmt.Errorf("failed to get initial database state: %w", err)
	}
	logger.Info("database up", "state", metrics)

	ticker := time.NewTicker(config.Heartbeat)

	// TODO: Go handle any jobs

	slog.Info("starting event loop")
	for {
		select {
		case <-ctx.Done():
			slog.Info("shutting down", "reason", ctx.Err().Error())
			return nil
		case _ = <-ticker.C:
			slog.Info("beat")
			// TODO: Put jobs on the queue
		}
	}
}
