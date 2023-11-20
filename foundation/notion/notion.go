package notion

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	apiKey        string
	version       string
	notionVersion string
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:        apiKey,
		version:       "v1",
		notionVersion: "2022-06-28",
	}
}

type NotionError struct {
	Object    string `json:"object"`
	Status    int    `json:"status"`
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestId string `json:"request_id"`
}

func (ne NotionError) Error() string {
	return ne.Message
}

func (c *Client) Request(method string, path string, body io.Reader, value any) error {
	url := fmt.Sprintf("https://api.notion.com/%s%s", c.version, path)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return fmt.Errorf("request contruction for url: %s: %w", url, err)
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Add("Notion-Version", c.notionVersion)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed for url: %s: %w", url, err)
	}

	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var ne NotionError
		if err := dec.Decode(&ne); err != nil {
			return fmt.Errorf("decode error body response: %w", err)
		}
		return fmt.Errorf("unsuccessful request with status code %d: %w", resp.StatusCode, ne)
	}

	if err := dec.Decode(value); err != nil {
		return fmt.Errorf("decode body response: %w", err)
	}

	return nil
}

type Blocks struct {
	Type string
}

func (nc *Client) BlockChildren(blockID string, blocks Blocks) error {
	if err := nc.Request(http.MethodGet, fmt.Sprintf("/blocks/%s/children?page_size=500", blockID), nil, &blocks); err != nil {
		return fmt.Errorf("retiriving block with id %s childern: %w", blockID, err)
	}

	return nil
}
