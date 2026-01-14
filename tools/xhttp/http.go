package xhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/Yet-Another-AI-Project/kiwi-lib/tools/otelutils"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// HTTPError represents an HTTP error, including status code and response body.
type HTTPError struct {
	StatusCode int
	Body       []byte
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("http request failed with status code %d, body: %s", e.StatusCode, string(e.Body))
}

type OtelTransport struct {
	Transport http.RoundTripper
}

// RoundTrip adds headers and delegates to the original Transport.
func (ot *OtelTransport) RoundTrip(req *http.Request) (*http.Response, error) {

	headers := otelutils.MapCarrier(req.Context())
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return ot.Transport.RoundTrip(req)
}

func NewOtelTransport(base http.RoundTripper) *OtelTransport {
	if base == nil {
		base = http.DefaultTransport
	}

	return &OtelTransport{
		Transport: otelhttp.NewTransport(base),
	}
}

type Client struct {
	timeout time.Duration
	http.Client
}

type ClientOption func(*Client)

func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.timeout = timeout
	}
}

func NewClient(opts ...ClientOption) *Client {
	httpClient := &Client{
		timeout: 5 * time.Second, // default timeout
	}
	for _, opt := range opts {
		opt(httpClient)
	}

	httpClient.Client = http.Client{
		Transport: NewOtelTransport(http.DefaultTransport),
		Timeout:   httpClient.timeout,
	}

	return httpClient
}

// Get sends a GET request. 'req' is converted to query parameters. 'res' is a pointer to unmarshal the response body.
func (c *Client) Get(ctx context.Context, url string, req, res any, header http.Header) error {
	return c.do(ctx, http.MethodGet, url, req, res, header)
}

// Post sends a POST request. 'req' is marshaled as JSON body. 'res' is a pointer to unmarshal the response body.
func (c *Client) Post(ctx context.Context, url string, req, res any, header http.Header) error {
	return c.do(ctx, http.MethodPost, url, req, res, header)
}

// Put sends a PUT request. 'req' is marshaled as JSON body. 'res' is a pointer to unmarshal the response body.
func (c *Client) Put(ctx context.Context, url string, req, res any, header http.Header) error {
	return c.do(ctx, http.MethodPut, url, req, res, header)
}

// Patch sends a PATCH request. 'req' is marshaled as JSON body. 'res' is a pointer to unmarshal the response body.
func (c *Client) Patch(ctx context.Context, url string, req, res any, header http.Header) error {
	return c.do(ctx, http.MethodPatch, url, req, res, header)
}

// Delete sends a DELETE request. 'req' is converted to query parameters.
func (c *Client) Delete(ctx context.Context, url string, req, res any, header http.Header) error {
	return c.do(ctx, http.MethodDelete, url, req, res, header)
}

func (c *Client) do(ctx context.Context, method, fullURL string, reqData, resData any, header http.Header) error {
	var bodyReader io.Reader

	// Handle request data based on method
	if method == http.MethodGet || method == http.MethodDelete {
		if reqData != nil {
			params, err := toURLValues(reqData)
			if err != nil {
				return fmt.Errorf("failed to convert request to query parameters: %w", err)
			}
			if len(params) > 0 {
				fullURL += "?" + params.Encode()
			}
		}
	} else { // POST, PUT, etc.
		if reqData != nil {
			bodyBytes, err := json.Marshal(reqData)
			if err != nil {
				return fmt.Errorf("failed to marshal request body: %w", err)
			}
			bodyReader = bytes.NewBuffer(bodyBytes)
		}
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if header != nil {
		httpReq.Header = header
	}

	if bodyReader != nil && httpReq.Header.Get("Content-Type") == "" {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.Client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &HTTPError{
			StatusCode: resp.StatusCode,
			Body:       respBody,
		}
	}

	if resData != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, resData); err != nil {
			return fmt.Errorf("failed to unmarshal response body: %w", err)
		}
	}

	return nil
}

// toURLValues converts an interface to url.Values for GET requests.
func toURLValues(data any) (url.Values, error) {
	values := url.Values{}
	if data == nil {
		return values, nil
	}

	switch v := data.(type) {
	case url.Values:
		return v, nil
	case map[string]string:
		for key, val := range v {
			values.Add(key, val)
		}
		return values, nil
	default: // Fallback to JSON marshaling for structs
		jsonBytes, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal data for query params: %w", err)
		}
		var m map[string]interface{}
		if err := json.Unmarshal(jsonBytes, &m); err != nil {
			return nil, fmt.Errorf("data for GET request must be a struct or map")
		}
		for key, val := range m {
			if val != nil {
				values.Set(key, fmt.Sprint(val))
			}
		}
		return values, nil
	}
}
