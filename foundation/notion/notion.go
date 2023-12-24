// Package notionprovider provides a client to call notionprovider API
package notion

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Client is a notionprovider client that can send requests to notionprovider API
type Client struct {
	apiKey        string
	version       string
	notionVersion string
}

const APIVersion = "v1"
const Version = "2022-06-28"
const Base = "https://api.notion.com"

// NewClient creates a new notionprovider client with default versions set
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:        apiKey,
		version:       APIVersion,
		notionVersion: Version,
	}
}

// NotionError represents an error that notionprovider API responded with when called
type NotionError struct {
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (ne NotionError) Error() string {
	return ne.Message
}

func (c *Client) Request(method string, path string, body io.Reader, value any) error {
	url := fmt.Sprintf("%s/%s%s", Base, c.version, path)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return fmt.Errorf("construct construct notionprovider request: url: %s: %w", url, err)
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Add("Notion-Version", c.notionVersion)
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("notionprovider request: url: %s: %w", url, err)
	}

	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	if resp.StatusCode >= http.StatusBadRequest {
		var ne NotionError
		if err := dec.Decode(&ne); err != nil {
			return fmt.Errorf("decode error body response: %w", err)
		}
		return fmt.Errorf("notionprovider request: status code %d: %w", resp.StatusCode, ne)
	}

	if err := dec.Decode(value); err != nil {
		return fmt.Errorf("decode body response: %w", err)
	}

	return nil
}
