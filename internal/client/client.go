// Package client provides a safe HTTP client for the labeling server.
//
// Security invariant: error messages and logs produced by this package
// must never include the server host, path, query string, or Authorization
// header values. All such details are stripped by safeNetErr.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/mjudcd-ct-r-d-labeling/labeling_download_cli/internal/build"
)

// ErrNotAuthorized is returned for HTTP 401 / 403 responses.
// All auth failures map to a single message so callers cannot distinguish
// between wrong credentials and insufficient role (SEC-007).
var ErrNotAuthorized = errors.New("Not authorized")

// ErrNotFound is returned for HTTP 404 responses.
var ErrNotFound = errors.New("file not found")

// Client wraps http.Client with Bearer auth and URL-safe error handling.
type Client struct {
	hc    *http.Client
	token string
}

// New returns an unauthenticated Client.
// ResponseHeaderTimeout prevents hanging on slow servers while allowing
// arbitrarily long response bodies (needed for video streaming).
func New() *Client {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.ResponseHeaderTimeout = 30 * time.Second
	return &Client{
		hc: &http.Client{Transport: t},
	}
}

// WithToken returns a new Client that attaches the given Bearer token.
func (c *Client) WithToken(token string) *Client {
	return &Client{hc: c.hc, token: token}
}

// PostJSON marshals body as JSON, POSTs to path, and decodes the response into out.
func (c *Client) PostJSON(ctx context.Context, path string, body any, out any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("request encoding error")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, build.Endpoint()+path, bytes.NewReader(data))
	if err != nil {
		return safeNetErr(err)
	}
	req.Header.Set("Content-Type", "application/json")
	c.attachAuth(req)

	resp, err := c.hc.Do(req)
	if err != nil {
		return safeNetErr(err)
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusUnauthorized,
		resp.StatusCode == http.StatusForbidden:
		return ErrNotAuthorized
	case resp.StatusCode < 200 || resp.StatusCode >= 300:
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

// GetJSON GETs path and decodes the JSON response into out.
func (c *Client) GetJSON(ctx context.Context, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, build.Endpoint()+path, nil)
	if err != nil {
		return safeNetErr(err)
	}
	c.attachAuth(req)

	resp, err := c.hc.Do(req)
	if err != nil {
		return safeNetErr(err)
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusUnauthorized,
		resp.StatusCode == http.StatusForbidden:
		return ErrNotAuthorized
	case resp.StatusCode < 200 || resp.StatusCode >= 300:
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// GetStream opens a streaming GET for path.
// When partOffset > 0, a Range header is set to attempt byte-range resume.
// The caller is responsible for closing the returned response body.
func (c *Client) GetStream(ctx context.Context, path string, partOffset int64) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, build.Endpoint()+path, nil)
	if err != nil {
		return nil, safeNetErr(err)
	}
	c.attachAuth(req)
	if partOffset > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", partOffset))
	}

	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, safeNetErr(err)
	}

	switch resp.StatusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		resp.Body.Close()
		return nil, ErrNotAuthorized
	case http.StatusNotFound:
		resp.Body.Close()
		return nil, ErrNotFound
	case http.StatusOK, http.StatusPartialContent:
		return resp, nil
	default:
		resp.Body.Close()
		return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}
}

// GetBytes fetches path without auth and returns the raw body.
func (c *Client) GetBytes(ctx context.Context, path string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, build.Endpoint()+path, nil)
	if err != nil {
		return nil, safeNetErr(err)
	}
	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, safeNetErr(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func (c *Client) attachAuth(req *http.Request) {
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
}

// safeNetErr strips all URL / host information from a network error so the
// server address is never written to the terminal or log files (SEC-004).
func safeNetErr(err error) error {
	if err == nil {
		return nil
	}
	// context cancellation is safe to propagate as-is
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return fmt.Errorf("Network error. Please try again.")
	}
	return fmt.Errorf("Network error. Please try again.")
}
