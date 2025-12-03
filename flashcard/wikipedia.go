package flashcard

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

const (
	FWikipediaSummaryLink = "https://en.wikipedia.org/api/rest_v1/page/summary/%s"
	FUserAgent            = "Flashcard_Bot/0.1 (%s) github.com/ohhfishal/fishy/0.1"
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
	if term.Wikipedia == "" {
		return nil, fmt.Errorf("cannot use wikipedia for: %v", term)
	}

	response, err := client.Do("GET", fmt.Sprintf(FWikipediaSummaryLink, term.Wikipedia), nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var summary WikipediaSummaryResponse
	if err := json.NewDecoder(response.Body).Decode(&summary); err != nil {
		return nil, fmt.Errorf("reading body: %w", err)
	}
	// TODO: Parse the actual html page to get more cards
	return []Flashcard{
		{
			Header:      term.Name,
			Description: summary.Extract,
			Origin:      "wikipedia",
		},
	}, nil
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
