package flashcard

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/goccy/go-yaml"
	"io"
	"log/slog"
)

type GenerateCMD struct {
	File       []byte         `arg:"" type:"filecontent" help:"YAML file describing terms to make flashcards of."`
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
	// TODO: Have this save them somewhere
	slog.Info("created flashcards", "flashcards", flashcards)
	if len(errs) > 0 {
		slog.Warn("got errors", "errs", errors.Join(errs...))
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
