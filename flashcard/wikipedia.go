package flashcard

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
)

const (
	FWikipediaSummaryLink  = "https://en.wikipedia.org/api/rest_v1/page/summary/%s"
	FWikipediaPageMarkdown = "[wikipedia](https://en.wikipedia.org/wiki/%s)"
	FUserAgent             = "Flashcard_Bot/0.1 (%s) github.com/ohhfishal/fishy/0.1"
)

type WikipediaSummaryResponse struct {
	Language string `json:"lang"`
	Title    string `json:"title"`
	Extract  string `json:"extract"`
}

type WikipediaClient struct {
	Contact string
}

func (client *WikipediaClient) CreateFlashcards(ctx context.Context, term Term) ([]Flashcard, error) {
	if len(term.Wikipedia) == 0 {
		return nil, fmt.Errorf("does not support wikipedia: %v", term)
	}

	var cards []Flashcard
	var errs error
	for _, article := range term.Wikipedia {
		card, err := client.CreateFlashcard(ctx, article, term.Name)
		errs = errors.Join(errs, err)
		if card != nil {
			cards = append(cards, *card)
		}
	}
	return cards, errs
}

func (client *WikipediaClient) CreateFlashcard(ctx context.Context, article string, header string) (*Flashcard, error) {
	var description string
	if strings.Contains(article, "#") {
		// TODO: Parse the actual html page
		return nil, fmt.Errorf("not implemented: headings: %s", article)
	} else {
		summary, err := Get[WikipediaSummaryResponse](client, fmt.Sprintf(FWikipediaSummaryLink, article), nil)
		if err != nil {
			return nil, err
		}
		description = summary.Extract
	}
	return &Flashcard{
		Header:      header,
		Description: description,
		Origin:      fmt.Sprintf(FWikipediaPageMarkdown, article),
	}, nil
}

func Get[T any](client *WikipediaClient, url string, body any) (T, error) {
	return Do[T](client, "GET", url, body)
}
func Do[T any](client *WikipediaClient, method string, url string, body any) (T, error) {
	var parsed T
	response, err := client.Do(method, url, body)
	if err != nil {
		return parsed, fmt.Errorf("requesting wikipedia: %w", err)
	}
	defer response.Body.Close()

	if err := json.NewDecoder(response.Body).Decode(&parsed); err != nil {
		return parsed, fmt.Errorf("reading body: %w", err)
	}
	return parsed, nil
}

func (client *WikipediaClient) Do(method string, url string, body any) (*http.Response, error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	request.Header.Set("User-Agent", fmt.Sprintf(FUserAgent, client.Contact))
	slog.Debug("performing get", "url", url, "headers", request.Header)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("making response to %s: %w", url, err)
	}
	if response.StatusCode >= 300 || response.StatusCode < 200 {
		slog.Warn("got", "status", response.Status)
		// TODO: HANDLE
	}
	// TODO: Slow down if we get a 429 etc
	return response, nil
}
