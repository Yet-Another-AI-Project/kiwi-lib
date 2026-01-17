package xhttp

import (
	"fmt"
	"net/http"
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
