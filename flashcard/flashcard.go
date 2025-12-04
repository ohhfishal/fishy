package flashcard

import (
	"context"
	"fmt"
	"github.com/goccy/go-yaml"
	"io"
	"os/exec"
	"strings"
)

type Flashcard struct {
	Header       string   `json:"header,omitempty"`
	Description  string   `json:"description,omitempty"`
	AIOverview   []string `json:"ai_overview,omitempty"`
	Origin       string   `json:"origin"`
	ClassContext string   `json:"class_context"`
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

type root struct {
	Textbooks []Textbook `json:"textbooks" yaml:"textbooks"`
}

type Textbook struct {
	Name     string    `json:"name" yaml:"name"`
	Subject  string    `json:"subject" yaml:"subject"`
	Chapters []Chapter `json:"chapters" yaml:"chapters"`
}

type Chapter struct {
	Number int    `json:"chapter" yaml"chapter"`
	Terms  []Term `json:"terms" yaml:"terms"`
}

type Term struct {
	Name      string   `json:"name" yaml:"name"`
	Passages  []string `json:"passages,omitempty" yaml:"passages"`
	Wikipedia []string `json:"wikipedia,omitempty" yaml:"wikipedia"`
}

func ParseTextbooks(ctx context.Context, reader io.Reader) ([]Textbook, error) {
	var root root
	decoder := yaml.NewDecoder(reader, yaml.DisallowUnknownField())
	if err := decoder.DecodeContext(ctx, &root); err != nil {
		return nil, fmt.Errorf("parsing yaml: %w", err)
	}
	return root.Textbooks, nil
}
