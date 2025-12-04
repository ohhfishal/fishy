package serve

import (
	"context"
	"log/slog"
	"time"
	"github.com/ohhfishal/fishy/notify"
)

type ServerConfig struct {
	Webhook      string       `arg:"" required:"" help:"Discord webhook to send message to."`
	Database string `arg:"" default:"fishy.db" help:"Path to SQLite file."`
	EmbedOptions notify.EmbedOptions `embed:""`

	CardFile string `name:"load" short:"l" type:"existingfile" help:"Generated flashcard file to load in. Ignores duplicates."`
	Interval time.Duration `default:"15m" help:"Duration between attempts to notify users."`
	Probability float64 `default:"0.50" help:"Starting probability a notification is send after interval."`
	Delta float64 `default:"0.1" help:"Delta added to probability on failure to trigger."`
}

type CMD struct {
	Config ServerConfig `embed:"" group:"Server Config"`
}

func (cmd *CMD) Run(ctx context.Context, logger *slog.Logger) error {
	if cmd.Config.CardFile != "" {
		logger.Warn("implement card loading", "file", cmd.Config.CardFile)
	}
	return nil
}
