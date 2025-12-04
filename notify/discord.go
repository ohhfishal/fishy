package notify

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"strings"

	"github.com/ohhfishal/fishy/discord"
	"github.com/ohhfishal/fishy/flashcard"
	"github.com/ohhfishal/fishy/version"
)

type DiscordCMD struct {
	Webhook      string       `arg:"" required:"" help:"Discord webhook to send message to."`
	File         *os.File     `default:"out.fish" help:"Fish file to load flashscard from."`
	EmbedOptions EmbedOptions `embed:"" group:"Embed Options"`
	DryRun       bool         `help:"Don't send the message and print it to stdout instead."`
}

func (config *DiscordCMD) Run(ctx context.Context, logger *slog.Logger) error {
	defer config.File.Close()

	var cards []flashcard.Flashcard
	decoder := json.NewDecoder(config.File)
	if err := decoder.Decode(&cards); err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	logger.Debug("read cards successfully", "total", len(cards))

	selected := cards[rand.Int()%len(cards)]
	embed := Embed(selected, config.EmbedOptions)
	if config.DryRun {
		fmt.Println(embed)
		return nil
	}
	slog.Info("sending", "embed", embed)
	if err := embed.Post(config.Webhook); err != nil {
		return fmt.Errorf("could not post embed: %v: %w", embed, err)
	}

	return nil

}

type EmbedOptions struct {
	Mentions []string `short:"m" help:"List of mentions to add."`
	// TODO: Add options
	// Spoil description
}

func Embed(card flashcard.Flashcard, opts EmbedOptions) discord.Embed {
	return discord.Embed{
		Content: strings.Join(opts.Mentions, " "),
		Messages: []discord.Message{
			{
				Title:       card.Header,
				Description: fmt.Sprintf("|| %s ||", card.Description),
				Color:       0x5865F2,
				Fields: []discord.Field{
					{
						Name:  "Source",
						Value: card.Origin,
					},
				},
				Footer: discord.Footer{
					Text: fmt.Sprintf("fishy %s • %s", version.Version(), version.Repo),
				},
			},
		},
	}
}
