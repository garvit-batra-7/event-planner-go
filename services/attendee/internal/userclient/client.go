// services/attendee/internal/userclient/clients.go
package userclient

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "net/http"
    "strings"
    "time"
)

var ErrNotFound = errors.New("user not found")

type Config struct {
    BaseURL string
    Timeout time.Duration
}

type User struct {
    ID       int    `json:"id"`
    Email    string `json:"email"`
    Name     string `json:"name"`
    Password string `json:"-"` 
}


type Client interface {
    GetUser(ctx context.Context, id int) (*User, error)
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

func (c *httpClient) GetUser(ctx context.Context, id int) (*User, error) {
    // Adjust path to match your User service API
    url := fmt.Sprintf("%s/api/v1/auth/users/%d", c.baseURL, id)

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
        var user User
        if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
            return nil, err
        }
        return &user, nil
    case http.StatusNotFound:
        return nil, ErrNotFound
    default:
        return nil, fmt.Errorf("user service returned status %d", resp.StatusCode)
    }
}
