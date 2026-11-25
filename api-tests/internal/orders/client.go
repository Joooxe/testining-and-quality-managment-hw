package orders

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const collectionPath = "/orders"

type Client struct {
	baseURL string
	http    *http.Client
}

type Response struct {
	StatusCode int
	Header     http.Header
	Body       []byte
	Path       string
}

func NewClient(baseURL string, client *http.Client) *Client {
	return &Client{baseURL: baseURL, http: client}
}

func (c *Client) Create(ctx context.Context, input NewOrder) (Response, error) {
	return c.request(ctx, http.MethodPost, collectionPath, input)
}

func (c *Client) CreateRawJSON(ctx context.Context, payload []byte) (Response, error) {
	return c.send(ctx, http.MethodPost, collectionPath, bytes.NewReader(payload))
}

func (c *Client) Get(ctx context.Context, id string) (Response, error) {
	return c.request(ctx, http.MethodGet, orderPath(id), nil)
}

func (c *Client) List(ctx context.Context) (Response, error) {
	return c.request(ctx, http.MethodGet, collectionPath, nil)
}

func (c *Client) ChangeStatus(ctx context.Context, id, status string) (Response, error) {
	payload := struct {
		Status string `json:"status"`
	}{Status: status}
	return c.request(ctx, http.MethodPatch, orderPath(id)+"/status", payload)
}

func (c *Client) Cancel(ctx context.Context, id string, reason *string) (Response, error) {
	path := orderPath(id)
	if reason != nil {
		path = withReason(path, *reason)
	}
	return c.request(ctx, http.MethodDelete, path, nil)
}

func (c *Client) Remove(ctx context.Context, id, reason string) (Response, error) {
	return c.request(ctx, http.MethodDelete, withReason(orderPath(id)+"/remove", reason), nil)
}

func orderPath(id string) string {
	return collectionPath + "/" + url.PathEscape(id)
}

func withReason(path, reason string) string {
	query := url.Values{"reason": {reason}}
	return path + "?" + query.Encode()
}

func (c *Client) request(ctx context.Context, method, path string, payload any) (Response, error) {
	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return Response{}, fmt.Errorf("encode %s %s: %w", method, path, err)
		}
		body = bytes.NewReader(data)
	}
	return c.send(ctx, method, path, body)
}

func (c *Client) send(ctx context.Context, method, path string, body io.Reader) (Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return Response{}, fmt.Errorf("build %s %s: %w", method, path, err)
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return Response{}, fmt.Errorf("%s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, fmt.Errorf("read %s %s: %w", method, path, err)
	}
	return Response{
		StatusCode: resp.StatusCode,
		Header:     resp.Header,
		Body:       data,
		Path:       req.URL.Path,
	}, nil
}
