package metahttp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/onmetahq/meta-http/pkg/models"
	"github.com/onmetahq/meta-http/pkg/utils"
)

type Client struct {
	BaseURL        string
	log            log.Logger
	HTTPClient     *http.Client
	defaultHeaders map[string]string
}

type Retry struct {
	MaxRetries        int
	DelayBetweenRetry time.Duration
	Validator         func(int) bool
}

func NewClient(baseUrl string, log log.Logger, timeout time.Duration) *Client {
	return &Client{
		BaseURL: baseUrl,
		log:     log,
		HTTPClient: &http.Client{
			Transport: &loggingRoundTripper{
				logger: log,
				next:   http.DefaultTransport,
			},
			Timeout: timeout,
		},
	}
}

func NewClientWithRetry(baseUrl string, log log.Logger, timeout time.Duration, retry Retry) *Client {
	return &Client{
		BaseURL: baseUrl,
		log:     log,
		HTTPClient: &http.Client{
			Transport: &retryRoundTripper{
				maxRetries: retry.MaxRetries,
				delay:      retry.DelayBetweenRetry,
				next: &loggingRoundTripper{
					logger: log,
					next:   http.DefaultTransport,
				},
				validator: retry.Validator,
			},
			Timeout: timeout,
		},
	}
}

func (c *Client) SetDefaultHeaders(headers map[string]string) {
	c.defaultHeaders = headers
}

type HttpClientErrorResponse struct {
	Success    bool      `json:"success"`
	Err        ErrorInfo `json:"error"`
	StatusCode int       `json:"_"`
}

type ErrorInfo struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (hce *HttpClientErrorResponse) Error() string {
	return fmt.Sprintf("StatusCode: %d, ErrorCode: %d, Message: %s", hce.StatusCode, hce.Err.Code, hce.Err.Message)
}

func (c *Client) sendRequest(req *http.Request, v interface{}) (*models.ResponseData, error) {
	response := models.ResponseData{}
	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return &response, err
	}

	response.Header = res.Header
	response.Status = res.Status
	response.StatusCode = res.StatusCode

	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		errRes := HttpClientErrorResponse{}
		errRes.StatusCode = res.StatusCode
		if err = json.NewDecoder(res.Body).Decode(&errRes); err == nil {
			return &response, &errRes
		}

		errRes.Err = ErrorInfo{
			Message: fmt.Sprintf("unknown error, status code: %d", res.StatusCode),
		}
		return &response, &errRes
	}

	if err = json.NewDecoder(res.Body).Decode(&v); err != nil {
		fmt.Println(err)
		return &response, err
	}
	return &response, nil
}

func generateUrl(basePath string, relativePath string) string {
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

func (c *Client) Get(ctx context.Context, path string, headers map[string]string, v interface{}) (*models.ResponseData, error) {
	response := models.ResponseData{}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, generateUrl(c.BaseURL, path), nil)
	if err != nil {
		return &response, err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json; charset=utf-8")

	for k, v := range c.defaultHeaders {
		req.Header.Set(k, v)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	ctxHeaders := utils.FetchHeadersFromContext(ctx)
	for k, v := range ctxHeaders {
		req.Header.Set(k, v)
	}

	resp, err := c.sendRequest(req, v)
	if err != nil {
		return &response, err
	}

	return resp, nil
}

func (c *Client) Post(ctx context.Context, path string, headers map[string]string, v interface{}, res interface{}) (*models.ResponseData, error) {
	response := models.ResponseData{}
	postBody, err := json.Marshal(v)
	if err != nil {
		return &response, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, generateUrl(c.BaseURL, path), bytes.NewBuffer(postBody))
	if err != nil {
		return &response, err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json; charset=utf-8")

	for k, v := range c.defaultHeaders {
		req.Header.Set(k, v)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	ctxHeaders := utils.FetchHeadersFromContext(ctx)
	for k, v := range ctxHeaders {
		req.Header.Set(k, v)
	}

	resp, err := c.sendRequest(req, res)
	if err != nil {
		return &response, err
	}

	return resp, nil
}

func (c *Client) Put(ctx context.Context, path string, headers map[string]string, v interface{}, res interface{}) (*models.ResponseData, error) {
	response := models.ResponseData{}
	postBody, err := json.Marshal(v)
	if err != nil {
		return &response, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, generateUrl(c.BaseURL, path), bytes.NewBuffer(postBody))
	if err != nil {
		return &response, err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json; charset=utf-8")

	for k, v := range c.defaultHeaders {
		req.Header.Set(k, v)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	ctxHeaders := utils.FetchHeadersFromContext(ctx)
	for k, v := range ctxHeaders {
		req.Header.Set(k, v)
	}

	resp, err := c.sendRequest(req, res)
	if err != nil {
		return &response, err
	}

	return resp, nil
}

type loggingRoundTripper struct {
	next   http.RoundTripper
	logger log.Logger
}

func (l loggingRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	presentTime := time.Now()
	//no lint
	level.Debug(l.logger).Log("msg", "Initiating call", "path", r.URL.Path, "host", r.URL.Host, string(models.RequestID), r.Header.Get(string(models.RequestID)))
	res, err := l.next.RoundTrip(r)
	if err != nil {
		level.Debug(l.logger).Log("msg", "Call Ended", "path", r.URL.Path, "host", r.URL.Host, "duration", time.Since(presentTime).Milliseconds(), "error", err.Error(), string(models.RequestID), r.Header.Get(string(models.RequestID)))
		return nil, err
	}
	level.Debug(l.logger).Log("msg", "Call Ended", "path", r.URL.Path, "host", r.URL.Host, "duration", time.Since(presentTime).Milliseconds(), "status", res.StatusCode, string(models.RequestID), r.Header.Get(string(models.RequestID)))
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
