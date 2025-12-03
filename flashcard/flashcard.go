package flashcard

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type Flashcard struct {
	Header      string `json:"header,omitempty"`
	Description string `json:"description,omitempty"`
	Origin      string `json:"origin"`
}

type FlashcardsArgs struct {
	Wikipedia WikipediaArgs `embed:"wikipedia" prefix:"wikipedia-" group:"Wikipedia"`
}

type WikipediaArgs struct {
	Disable   bool   `help:"Don't generate cards using wikipedia."`
	UserAgent string `help:"User-Agent field in making requests. If empty uses 'git config user.email'."`
}

func (args *WikipediaArgs) AfterApply(ctx context.Context) error {
	if args.Disable {
		return nil
	}
	if args.UserAgent == "" {
		rawBytes, err := exec.CommandContext(ctx, "git", "config", "user.email").Output()
		if err != nil {
			return fmt.Errorf("could not infer user agent: %w", err)
		}
		args.UserAgent = strings.TrimSpace(string(rawBytes))
	}
	return nil
}
