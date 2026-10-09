// Package imgbb uploads images to imgbb.com and returns the hosted URL. It is
// used for profile pictures so the database only stores a URL, never image
// bytes.
package imgbb

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

const (
	defaultEndpoint = "https://api.imgbb.com/1/upload"
	requestTimeout  = 20 * time.Second
)

// ErrNotConfigured is returned when no API key is set.
var ErrNotConfigured = errors.New("imgbb: IMGBB_API_KEY is not configured")

// Client uploads images to imgbb.com.
type Client struct {
	apiKey   string
	endpoint string
	http     *http.Client
}

// New builds a client. An empty apiKey still yields a usable client whose
// Upload returns ErrNotConfigured.
func New(apiKey string) *Client {
	return &Client{
		apiKey:   apiKey,
		endpoint: defaultEndpoint,
		http:     &http.Client{Timeout: requestTimeout},
	}
}

// Upload sends the image bytes and returns the hosted image URL.
func (c *Client) Upload(ctx context.Context, filename string, data []byte) (string, error) {
	if c.apiKey == "" {
		return "", ErrNotConfigured
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("image", filename)
	if err != nil {
		return "", err
	}
	if _, err := part.Write(data); err != nil {
		return "", err
	}
	if err := writer.Close(); err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint+"?key="+c.apiKey, &body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", err
	}

	var parsed struct {
		Success bool `json:"success"`
		Data    struct {
			URL        string `json:"url"`
			DisplayURL string `json:"display_url"`
		} `json:"data"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", fmt.Errorf("imgbb: unexpected response (status %d)", resp.StatusCode)
	}
	if !parsed.Success || parsed.Data.URL == "" {
		msg := parsed.Error.Message
		if msg == "" {
			msg = fmt.Sprintf("upload failed (status %d)", resp.StatusCode)
		}
		return "", fmt.Errorf("imgbb: %s", msg)
	}

	return parsed.Data.URL, nil
}
