package serve

import (
	"context"
	"errors"
	"fmt"
	"github.com/ohhfishal/fishy/database"
	"github.com/ohhfishal/fishy/notify"
	"log/slog"
	"math/rand"
	"time"
)

var ErrSkip = errors.New("no work to do")

type CMD struct {
	Config ServerConfig `embed:"" group:"Server Config"`
}

type ServerConfig struct {
	Webhook      string              `arg:"" required:"" help:"Discord webhook to send message to."`
	Database     string              `arg:"" default:"fishy.db" help:"SQLite connection string."`
	EmbedOptions notify.EmbedOptions `embed:""`
	CardFile     string              `name:"load" short:"l" type:"existingfile" help:"Generated flashcard file to load in. Ignores duplicates."`
	Interval     time.Duration       `default:"15m" help:"Minimum duration between notifications."`
	Heartbeat    time.Duration       `default:"1m" help:"Duration between checks if there is work to be done."`
	Probability  float64             `default:"0.50" help:"Starting probability a notification is send after interval."`
	Delta        float64             `default:"0.1" help:"Delta added to probability on failure to trigger."`
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
		if results, err := db.LoadFlashcardsFrom(ctx, config.CardFile); err != nil {
			return fmt.Errorf("failed to load cards: %w", err)
		} else {
			logger.Info("loaded cards", "cards", results)
		}
	}

	metrics, err := db.Metrics(ctx)
	if err != nil {
		return fmt.Errorf("failed to get initial database state: %w", err)
	}
	logger.Info("database up", "state", metrics)

	ticker := time.NewTicker(config.Heartbeat)

	// Handle any jobs that are ready to run
	config.TickIfReady(ctx, db, logger)

	slog.Info("starting event loop")
	for {
		select {
		case <-ctx.Done():
			slog.Info("shutting down", "reason", ctx.Err().Error())
			return nil
		case _ = <-ticker.C:
			go config.TickIfReady(ctx, db, logger)
		}
	}
}

func (config *ServerConfig) TickIfReady(ctx context.Context, db *database.Store, logger *slog.Logger) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	logger = logger.With("job", "work")

	if err := config.ready(ctx, db, logger); err != nil {
		if !errors.Is(ErrSkip, err) {
			logger.Error("determining if ready", "err", err)
		}
		return
	}

	if err := config.Tick(ctx, db, logger); err != nil {
		logger.Error("error doing work", "err", err)
	}
}
func (config *ServerConfig) Tick(ctx context.Context, db *database.Store, logger *slog.Logger) error {
	// Pick a card.
	cards, err := db.GetCards(ctx)
	if err != nil {
		return fmt.Errorf("getting cards: %w", err)
	}

	// Do the notification stuff
	selected := database.ConvertFlashcard(cards[rand.Int()%len(cards)])
	embed := notify.Embed(selected, config.EmbedOptions)
	if err := embed.Post(config.Webhook); err != nil {
		return fmt.Errorf("could not post embed: %v: %w", embed, err)
	}

	// Put a new job
	// TODO: We don't want this operation to timeout
	job, err := db.PutJob(context.TODO(), 0)
	if err != nil {
		// This one is really bad since we might start thrashing and always send response
		return fmt.Errorf("inserting job", "err", err)
	}
	logger.Info("inserted", "job", job)
	return nil
}

// Return nil: Ready, ErrSkip: Not Ready, Err: Actual error
func (config *ServerConfig) ready(ctx context.Context, db *database.Store, logger *slog.Logger) error {
	jobs, err := db.GetLastJob(ctx)
	if err != nil {
		return fmt.Errorf("getting last job: %w", err)
	} else if len(jobs) == 0 {
		return nil
	}

	job := jobs[0]
	if time.Since(job.CreatedAt) < config.Interval {
		logger.Debug("not ready (time)")
		return ErrSkip
	}

	target := config.Probability + (config.Delta * float64(job.Failures))
	roll := rand.Float64()
	logger.Info("rolling", "target", target, "roll", roll, "status", roll >= target, "job", job)
	if roll >= target {
		return nil
	}

	// Log that we failed
	// TODO: We don't want this operation to timeout
	_, err = db.PutJob(context.TODO(), job.Failures+1)
	if err != nil {
		return err
	}
	return ErrSkip
}
