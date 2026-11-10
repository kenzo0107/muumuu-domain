package muumuu

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	BaseURLProduction = "https://muumuu-domain.com/api/v2"
	BaseURLSandbox    = "https://api-sandbox.muumuu-domain.com/api/v2"

	defaultUserAgent = "kenzo0107/muumuu-domain"
)

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

type Client struct {
	token      string
	baseURL    *url.URL
	userAgent  string
	httpClient HTTPClient
}

type Option func(*Client) error

func OptionBaseURL(endpoint string) Option {
	return func(c *Client) error {
		u, err := url.Parse(strings.TrimRight(endpoint, "/"))
		if err != nil {
			return fmt.Errorf("invalid base URL %q: %w", endpoint, err)
		}
		c.baseURL = u
		return nil
	}
}

func OptionHTTPClient(client HTTPClient) Option {
	return func(c *Client) error {
		if client == nil {
			return errors.New("http client must be non-nil")
		}
		c.httpClient = client
		return nil
	}
}

func OptionUserAgent(ua string) Option {
	return func(c *Client) error {
		c.userAgent = ua
		return nil
	}
}

func New(token string, options ...Option) (*Client, error) {
	baseURL, err := url.Parse(BaseURLProduction)
	if err != nil {
		return nil, err
	}

	c := &Client{
		token:      token,
		baseURL:    baseURL,
		userAgent:  defaultUserAgent,
		httpClient: http.DefaultClient,
	}

	for _, opt := range options {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	return c, nil
}

func (c *Client) newRequest(ctx context.Context, method, path string, query url.Values, body any) (*http.Request, error) {
	rawPath := c.baseURL.EscapedPath() + path
	unescaped, err := url.PathUnescape(rawPath)
	if err != nil {
		return nil, err
	}
	u := *c.baseURL
	u.Path = unescaped
	u.RawPath = rawPath
	u.RawQuery = query.Encode()

	var buf io.Reader
	if body != nil {
		b := &bytes.Buffer{}
		enc := json.NewEncoder(b)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(body); err != nil {
			return nil, err
		}
		buf = b
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

func (c *Client) do(req *http.Request, v any) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		return newError(resp)
	}

	if v == nil {
		return nil
	}

	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}
	return nil
}

type envelope[T any] struct {
	Data T               `json:"data"`
	Meta *PaginationMeta `json:"meta,omitempty"`
}

func call[T any](ctx context.Context, c *Client, method, path string, query url.Values, body any) (*envelope[T], error) {
	req, err := c.newRequest(ctx, method, path, query, body)
	if err != nil {
		return nil, err
	}

	r := new(envelope[T])
	if err := c.do(req, r); err != nil {
		return nil, err
	}
	return r, nil
}

func (c *Client) callNoContent(ctx context.Context, method, path string) error {
	req, err := c.newRequest(ctx, method, path, nil, nil)
	if err != nil {
		return err
	}
	return c.do(req, nil)
}

type PaginationMeta struct {
	Total    int `json:"total"`
	Page     int `json:"page"`
	PageSize int `json:"page-size"`
}

func (m PaginationMeta) HasNext() bool {
	return m.PageSize > 0 && m.Page*m.PageSize < m.Total
}

type ListOptions struct {
	Page     int
	PageSize int
}

func (o ListOptions) setTo(q url.Values) {
	setInt(q, "page", o.Page)
	setInt(q, "page-size", o.PageSize)
}

func setString(q url.Values, key, value string) {
	if value != "" {
		q.Set(key, value)
	}
}

func setInt(q url.Values, key string, value int) {
	if value != 0 {
		q.Set(key, fmt.Sprint(value))
	}
}

func setIntPtr(q url.Values, key string, value *int) {
	if value != nil {
		q.Set(key, fmt.Sprint(*value))
	}
}

func Bool(v bool) *bool { return &v }

func Int(v int) *int { return &v }

func String(v string) *string { return &v }
