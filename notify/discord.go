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
	File         string       `default:"out.json" type:"existingfile" help:"Fish file to load flashscard from."`
	EmbedOptions EmbedOptions `embed:"" group:"Embed Options"`
	DryRun       bool         `help:"Don't send the message and print it to stdout instead."`
}

func (config *DiscordCMD) Run(ctx context.Context, logger *slog.Logger) error {
	file, err := os.Open(config.File)
	if err != nil {
		return fmt.Errorf("opening file %s: %w", config.File, err)
	}
	defer file.Close()

	var cards []flashcard.Flashcard
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cards); err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	logger.Debug("read cards successfully", "total", len(cards))

	selected := cards[rand.Int()%len(cards)]
	embed := Embed(selected, config.EmbedOptions)
	slog.Info("sending", "embed", embed)
	if config.DryRun {
		return nil
	}
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
	fields := []discord.Field{
		{
			Name:  "Source",
			Value: card.Origin,
		},
	}
	if card.ClassContext != "" {
		fields = append(fields, discord.Field{
			Name:  "Textbook",
			Value: card.ClassContext,
		})
	}
	if len(card.AIOverview) > 0 {
		fields = append(fields, discord.Field{
			Name:  "AI Summary",
			Value: ConvertToBullets(card.AIOverview),
		})
	}
	var thumbnail discord.Image
	if card.Thumbnail.Source != "" {
		thumbnail = discord.Image{
					URL: card.Thumbnail.Source,
					Width: card.Thumbnail.Width,
					Height: card.Thumbnail.Height,
		}
	}
	return discord.Embed{
		Content: strings.Join(opts.Mentions, " "),
		Messages: []discord.Message{
			{
				Title:       card.Header,
				Description: fmt.Sprintf("||%s||", card.Description),
				Color:       0x5865F2,
				Fields:      fields,
				Footer: discord.Footer{
					Text: fmt.Sprintf("fishy %s • %s", version.Version(), version.Repo),
				},
				Image: thumbnail,
			},
		},
	}
}

func ConvertToBullets(lines []string) string {
	var builder strings.Builder
	for _, line := range lines {
		builder.WriteString(fmt.Sprintf("- ||%s||\n", line))
	}
	return strings.TrimSpace(builder.String())
}
