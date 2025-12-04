package flashcard

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/goccy/go-yaml"
	"io"
	"log/slog"
	"os"
)

type GenerateCMD struct {
	File       []byte         `arg:"" type:"filecontent" help:"YAML file describing terms to make flashcards of."`
	Output     string         `short:"o" default:"out.fish" type:"path" help:"File to write to."`
	Flashcards FlashcardsArgs `embed:""`
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
	Passages  []string `json:"passages,omitzero,omitempty" yaml:"passages"`
	Wikipedia string   `json:"wikipedia,omitempty" yaml:"wikipedia"`
}

func (config *GenerateCMD) Run(ctx context.Context, logger *slog.Logger) error {
	textbooks, err := ParseTextbooks(ctx, bytes.NewReader(config.File))
	if err != nil {
		return err
	}
	slog.Debug("got", "textbooks", textbooks)

	wikipediaClient := &WikipediaClient{
		Contact: config.Flashcards.Wikipedia.UserAgent,
	}

	var flashcards []Flashcard
	var errs []error
	for _, textbook := range textbooks {
		for _, chapter := range textbook.Chapters {
			for _, term := range chapter.Terms {
				if !config.Flashcards.Wikipedia.Disable {
					wikipedia, err := wikipediaClient.CreateFlashcards(ctx, term)
					if err != nil {
						errs = append(errs, fmt.Errorf("wikipedia: %w", err))
					} else {
						flashcards = append(flashcards, wikipedia...)
					}
				}
			}
		}
	}

	if len(errs) > 0 {
		msgs := []string{}
		for _, err := range errs {
			msgs = append(msgs, err.Error())
		}
		slog.Warn("got errors", "errs", msgs)
	}

	// Write to file
	file, err := os.OpenFile(config.Output, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("opening output file: %w", err)
	}

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(flashcards); err != nil {
		return fmt.Errorf("writing to output: %w", err)
	}

	return nil
}

func NewFlashcardsFor(ctx context.Context, term Term, features FlashcardsArgs) ([]Flashcard, []error) {
	var cards []Flashcard
	var errs []error

	return cards, errs
}

func ParseTextbooks(ctx context.Context, reader io.Reader) ([]Textbook, error) {
	var root root
	decoder := yaml.NewDecoder(reader, yaml.DisallowUnknownField())
	if err := decoder.DecodeContext(ctx, &root); err != nil {
		return nil, fmt.Errorf("parsing yaml: %w", err)
	}
	return root.Textbooks, nil
}
