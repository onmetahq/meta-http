package metahttp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/onmetahq/meta-http/pkg/models"
	"github.com/onmetahq/meta-http/pkg/utils"
)

type RequestOptions struct {
	URL     string
	Timeout time.Duration
}

type Requests interface {
	SetDefaultHeaders(headers map[string]string)
	Get(ctx context.Context, path string, headers map[string]string, v interface{}) (*models.ResponseData, error)
	Post(ctx context.Context, path string, headers map[string]string, v interface{}, res interface{}) (*models.ResponseData, error)
	Put(ctx context.Context, path string, headers map[string]string, v interface{}, res interface{}) (*models.ResponseData, error)
	GetConfig() RequestOptions
}

type client struct {
	BaseURL        string
	HTTPClient     *http.Client
	defaultHeaders map[string]string
}

func NewClient(baseUrl string, log *slog.Logger, timeout time.Duration) Requests {
	return &client{
		BaseURL: baseUrl,
		HTTPClient: &http.Client{
			Transport: &loggingRoundTripper{
				logger: log,
				next:   defaultPooledTransport(),
			},
			Timeout: timeout,
		},
	}
}

func NewClientWithRetry(baseUrl string, log *slog.Logger, timeout time.Duration, retry models.Retry) Requests {
	return &client{
		BaseURL: baseUrl,
		HTTPClient: &http.Client{
			Transport: &retryRoundTripper{
				maxRetries: retry.MaxRetries,
				delay:      retry.DelayBetweenRetry,
				next: &loggingRoundTripper{
					logger: log,
					next:   defaultPooledTransport(),
				},
				validator: retry.Validator,
			},
			Timeout: timeout,
		},
	}
}

func (c *client) SetDefaultHeaders(headers map[string]string) {
	c.defaultHeaders = headers
}

func (c *client) sendRequest(req *http.Request, v interface{}) (*models.ResponseData, error) {
	response := models.ResponseData{}
	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	response.Header = res.Header
	response.Status = res.Status
	response.StatusCode = res.StatusCode

	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		errRes := models.HttpClientErrorResponse{}
		errRes.StatusCode = res.StatusCode
		if err = json.NewDecoder(res.Body).Decode(&errRes); err == nil {
			return &response, &errRes
		}

		body, _ := io.ReadAll(res.Body)
		errRes.Err = models.ErrorInfo{
			Message: fmt.Sprintf("unknown error, status code: %d, response: %s", res.StatusCode, string(body)),
		}
		return &response, &errRes
	}

	if err = json.NewDecoder(res.Body).Decode(&v); err != nil {
		errRes := models.HttpClientErrorResponse{}
		errRes.Success = false
		errRes.StatusCode = http.StatusInternalServerError
		errRes.Err.Message = err.Error()
		return &response, &errRes
	}
	return &response, nil
}

func generateUrl(basePath string, relativePath string) string {
	if len(basePath) == 0 {
		return relativePath
	}

	x := basePath[len(basePath)-1]
	if x == '/' {
		if relativePath == "" {
			return basePath[:len(basePath)-1]
		} else if relativePath[0] == '/' {
			return basePath[:len(basePath)-1] + relativePath
		} else {
			return basePath + relativePath
		}
	} else {
		if relativePath == "" {
			return basePath
		} else if relativePath[0] == '/' {
			return basePath + relativePath
		} else {
			return basePath + "/" + relativePath
		}
	}
}

func (c *client) Get(ctx context.Context, path string, headers map[string]string, v interface{}) (*models.ResponseData, error) {
	ul := generateUrl(c.BaseURL, path)
	u, err := url.ParseRequestURI(ul)
	if err != nil || u.Host == "" || u.Scheme == "" {
		return nil, fmt.Errorf("%w url: %s, err: %v", models.ErrBadURL, ul, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ul, nil)
	if err != nil {
		return nil, err
	}

	ctxHeaders := utils.FetchHeadersFromContext(ctx)
	for k, v := range ctxHeaders {
		req.Header.Set(k, v)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json; charset=utf-8")

	for k, v := range c.defaultHeaders {
		req.Header.Set(k, v)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return c.sendRequest(req, v)
}

func (c *client) Post(ctx context.Context, path string, headers map[string]string, v interface{}, res interface{}) (*models.ResponseData, error) {
	ul := generateUrl(c.BaseURL, path)
	u, err := url.ParseRequestURI(ul)
	if err != nil || u.Host == "" || u.Scheme == "" {
		return nil, fmt.Errorf("invalid url, url: %s, err: %v", ul, err)
	}

	postBody, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ul, bytes.NewBuffer(postBody))
	if err != nil {
		return nil, err
	}

	ctxHeaders := utils.FetchHeadersFromContext(ctx)
	for k, v := range ctxHeaders {
		req.Header.Set(k, v)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json; charset=utf-8")

	for k, v := range c.defaultHeaders {
		req.Header.Set(k, v)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return c.sendRequest(req, res)
}

func (c *client) Put(ctx context.Context, path string, headers map[string]string, v interface{}, res interface{}) (*models.ResponseData, error) {
	ul := generateUrl(c.BaseURL, path)
	u, err := url.ParseRequestURI(ul)
	if err != nil || u.Host == "" || u.Scheme == "" {
		return nil, fmt.Errorf("invalid url, url: %s, err: %v", ul, err)
	}

	postBody, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, ul, bytes.NewBuffer(postBody))
	if err != nil {
		return nil, err
	}

	ctxHeaders := utils.FetchHeadersFromContext(ctx)
	for k, v := range ctxHeaders {
		req.Header.Set(k, v)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json; charset=utf-8")

	for k, v := range c.defaultHeaders {
		req.Header.Set(k, v)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return c.sendRequest(req, res)
}

func (c *client) GetConfig() RequestOptions {
	return RequestOptions{
		URL:     c.BaseURL,
		Timeout: c.HTTPClient.Timeout,
	}
}

type loggingRoundTripper struct {
	next   http.RoundTripper
	logger *slog.Logger
}

func (l loggingRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	presentTime := time.Now()
	l.logger.Debug(
		"Initiating call",
		slog.String("path", r.URL.Path),
		slog.String("host", r.URL.Host),
		slog.String(string(models.RequestID), r.Header.Get(string(models.RequestID))),
	)
	res, err := l.next.RoundTrip(r)
	if err != nil {
		l.logger.Debug(
			"Call Ended",
			slog.String("path", r.URL.Path),
			slog.String("host", r.URL.Host),
			slog.Int64("duration", time.Since(presentTime).Milliseconds()),
			slog.Any("error", err.Error()),
			slog.String(string(models.RequestID), r.Header.Get(string(models.RequestID))),
		)
		return nil, err
	}
	l.logger.Debug(
		"Call Ended",
		slog.String("path", r.URL.Path),
		slog.String("host", r.URL.Host),
		slog.Int64("duration", time.Since(presentTime).Milliseconds()),
		slog.Int("status", res.StatusCode),
		slog.String(string(models.RequestID), r.Header.Get(string(models.RequestID))),
	)
	return res, err
}

type retryRoundTripper struct {
	next       http.RoundTripper
	maxRetries int
	delay      time.Duration
	validator  func(int) bool
}

func (rrt retryRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	attempts := 0
	for {
		res, err := rrt.next.RoundTrip(r)
		attempts = attempts + 1

		if attempts == rrt.maxRetries {
			return res, err
		}

		if err == nil && rrt.validator(res.StatusCode) {
			return res, err
		}

		select {
		case <-r.Context().Done():
			return res, r.Context().Err()
		case <-time.After(rrt.delay):
		}
	}
}
