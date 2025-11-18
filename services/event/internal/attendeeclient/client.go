package attendeeclient

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Config struct {
	BaseURL string
	Timeout time.Duration
}

type Client interface {
	AddAttendee(ctx context.Context, eventID, userID int, authHeader string) error
	DeleteAttendeesForEvent(ctx context.Context, eventID int, authHeader string) error
}

type httpClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewHTTPClient(cfg Config) Client {
	if cfg.Timeout <= 0 {
		cfg.Timeout = 5 * time.Second
	}
	return &httpClient{
		baseURL: strings.TrimRight(cfg.BaseURL, "/"),
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

func (c *httpClient) AddAttendee(ctx context.Context, eventID, userID int, authHeader string) error {
	url := fmt.Sprintf("%s/api/v1/events/%d/attendees/%d", c.baseURL, eventID, userID)

	// 🔍 DEBUG LOG
	fmt.Printf("\n[attendeeclient] POST %s\n", url)
	if authHeader == "" {
		fmt.Println("[attendeeclient] ❌ No Authorization header provided!")
	} else {
		fmt.Printf("[attendeeclient] ✔ Forwarding Authorization header: %s...\n", authHeader[:20])
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return err
	}

	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		fmt.Printf("[attendeeclient] ❌ HTTP request failed: %v\n", err)
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("[attendeeclient] ⬅ Response status: %d\n\n", resp.StatusCode)

	if resp.StatusCode >= 300 {
		return fmt.Errorf("attendee service returned status %d", resp.StatusCode)
	}
	return nil
}

func (c *httpClient) DeleteAttendeesForEvent(ctx context.Context, eventID int, authHeader string) error {
	url := fmt.Sprintf("%s/api/v1/events/%d/attendees", c.baseURL, eventID)

	// 🔍 DEBUG LOG
	fmt.Printf("\n[attendeeclient] DELETE %s\n", url)
	if authHeader == "" {
		fmt.Println("[attendeeclient] ❌ No Authorization header provided!")
	} else {
		fmt.Printf("[attendeeclient] ✔ Forwarding Authorization header: %s...\n", authHeader[:20])
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		fmt.Printf("[attendeeclient] ❌ HTTP request failed: %v\n", err)
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("[attendeeclient] ⬅ Response status: %d\n\n", resp.StatusCode)

	if resp.StatusCode >= 300 {
		return fmt.Errorf("attendee service returned status %d", resp.StatusCode)
	}
	return nil
}
