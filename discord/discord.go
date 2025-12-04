package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Embed struct {
	Content         string          `json:"content,omitzero"`
	AllowedMentions AllowedMentions `json:"allowed_mentions,omitzero"`
	Messages        []Message       `json:"embeds,omitzero"`
}

type AllowedMentions struct {
	Parse []string `json:"parse"`
	Users []string `json:"users"`
}

type Footer struct {
	Text string `json:"text"`
}

type Author struct {
	Name string `json:"name"`
	Url  string `json:"url,omitzero"`
}

type Message struct {
	Title       string  `json:"title"`
	Description string  `json:"description,omitzero"`
	Color       int     `json:"color,omitzero"`
	Fields      []Field `json:"fields,omitempty"`
	Footer      Footer  `json:"footer,omitzero"`
	Author      Author  `json:"author,omitzero"`
	Image Image `json:"image,omitzero"`
}

type Image struct {
	URL string `json:"url"`
	Height int `json:"height,omitzero"`
	Width int `json:"width,omitzero"`
}

type Field struct {
	Name    string `json:"name"`
	Value   string `json:"value"`
	Inline  bool   `json:"inline"`
	TimeISO string `json:"timestamp,omitzero"`
}

func (embed *Embed) Post(webhook string) error {
	data, err := json.Marshal(embed)
	if err != nil {
		return err
	}

	request, err := http.NewRequest("POST", webhook, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	content, _ := io.ReadAll(response.Body)

	if response.StatusCode >= 400 {
		return fmt.Errorf("response failed: got: %d: %s", response.StatusCode, string(content))
	}
	return nil
}
