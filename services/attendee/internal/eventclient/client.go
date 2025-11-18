// services/attendee/internal/eventclient/clients.go
package eventclient

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "net/http"
    "strings"
    "time"
)

var ErrNotFound = errors.New("event not found")

type Config struct {
    BaseURL string
    Timeout time.Duration
}

// Event is what we get from the Event service.
// Adjust fields/tags to your real API.
type Event struct {
    ID          int    `json:"id"`
    OwnerID     int    `json:"ownerId"`
    Name        string `json:"name" binding:"required"`
    Description string `json:"description" binding:"required,min=10"`
    Date        string `json:"date" binding:"required,datetime=2006-01-02|datetime=2006-01-02T15:04:05Z07:00"`
    Location    string `json:"location" binding:"required,min=3"`
}

type Client interface {
    GetEvent(ctx context.Context, id int) (*Event, error)
}

type httpClient struct {
    baseURL string
    http    *http.Client
}

func NewHTTPClient(cfg Config) Client {
    return &httpClient{
        baseURL: strings.TrimRight(cfg.BaseURL, "/"),
        http: &http.Client{
            Timeout: cfg.Timeout,
        },
    }
}

func (c *httpClient) GetEvent(ctx context.Context, id int) (*Event, error) {
    url := fmt.Sprintf("%s/api/v1/events/%d", c.baseURL, id) // adjust path if needed

    req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    if err != nil {
        return nil, err
    }

    resp, err := c.http.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    switch resp.StatusCode {
    case http.StatusOK:
        var evt Event
        if err := json.NewDecoder(resp.Body).Decode(&evt); err != nil {
            return nil, err
        }
        return &evt, nil
    case http.StatusNotFound:
        return nil, ErrNotFound
    default:
        return nil, fmt.Errorf("event service returned status %d", resp.StatusCode)
    }
}
