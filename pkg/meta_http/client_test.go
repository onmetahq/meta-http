package metahttp_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	metahttp "github.com/onmetahq/meta-http/pkg/meta_http"
	"github.com/onmetahq/meta-http/pkg/models"
	"github.com/onmetahq/meta-http/pkg/utils"
)

func TestMetaHTTPClient(t *testing.T) {
	responseBody := "{\"Goodbye\":\"World\"}"
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte(responseBody))
	}))
	defer server.Close()

	logLevel := &slog.LevelVar{} // INFO
	logLevel.Set(slog.LevelDebug)
	opts := &slog.HandlerOptions{
		AddSource: true,
		Level:     logLevel,
	}

	logger := slog.New(slog.NewJSONHandler(os.Stderr, opts))

	metaHttpClient := metahttp.NewClient(server.URL, logger, 10*time.Second)
	req := struct {
		Hello string
	}{
		Hello: "world",
	}
	var res struct {
		Goodbye string
	}
	resp, err := metaHttpClient.Post(context.Background(), "/test", map[string]string{}, req, &res)
	if err != nil {
		t.Error(err.Error())
	}
	if res.Goodbye != "World" {
		t.Error("Response body is not as expected")
	}
	if resp != nil {
		t.Log(resp.StatusCode)
	}
}

func TestTimeoutScenario(t *testing.T) {
	responseBody := "{\"Goodbye\":\"World\"}"
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(2 * time.Second)
		rw.Write([]byte(responseBody))
	}))
	defer server.Close()

	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	metaHttpClient := metahttp.NewClient(server.URL, logger, 1*time.Second)
	req := struct {
		Hello string
	}{
		Hello: "world",
	}
	var res struct {
		Goodbye string
	}
	resp, err := metaHttpClient.Post(context.Background(), "/test", map[string]string{}, req, &res)
	if err == nil {
		t.Error("Supposed to fail with error")
	}
	if resp != nil {
		t.Log(resp.Status)
	}
}

func TestContextHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		userId := req.Header.Get(string(models.UserID))
		res := map[string]string{
			string(models.UserID): userId,
		}
		bytes, _ := json.Marshal(res)
		rw.Write(bytes)
	}))
	defer server.Close()

	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	metaHttpClient := metahttp.NewClient(server.URL, logger, 1*time.Second)
	req := struct {
		Hello string
	}{
		Hello: "world",
	}
	var res map[string]string
	ctx := context.Background()
	ctx = context.WithValue(ctx, models.UserID, "userId")
	ctx = context.WithValue(ctx, models.RequestID, "request-id")

	resp, err := metaHttpClient.Post(ctx, "/test", map[string]string{}, req, &res)
	if err != nil {
		t.Error(err.Error())
	}
	if res[string(models.UserID)] != "userId" {
		t.Error("Response body is not as expected")
	}
	if resp != nil {
		t.Log(resp.StatusCode)
	}
}

func TestHeadersContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		res := map[string]string{}
		ctx := utils.FetchContextFromHeaders(context.Background(), req)
		if userId, ok := ctx.Value(models.UserID).(string); ok {
			res["data"] = userId
		}
		if forwardedFor, ok := ctx.Value(models.XForwardedFor).(string); ok {
			res["forward"] = forwardedFor
		}
		bytes, _ := json.Marshal(res)
		rw.Write(bytes)
	}))
	defer server.Close()

	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	metaHttpClient := metahttp.NewClient(server.URL, logger, 1*time.Second)
	req := struct {
		Hello string
	}{
		Hello: "world",
	}
	var res map[string]string
	ctx := context.Background()

	headers := map[string]string{
		string(models.UserID):        "abcd",
		string(models.XForwardedFor): "10.0.9.8",
	}

	resp, err := metaHttpClient.Post(ctx, "/test", headers, req, &res)
	if err != nil {
		t.Error(err.Error())
	}

	if res["data"] != "abcd" {
		t.Error("Response body is not as expected")
	}
	if res["forward"] != "10.0.9.8" {
		t.Error("Response body is not as expected")
	}

	if resp != nil {
		t.Log(resp.Status)
	}
	res = map[string]string{}
	resp, err = metaHttpClient.Post(ctx, "/test", map[string]string{}, req, &res)
	if err != nil {
		t.Error(err.Error())
	}
	if _, ok := res["data"]; ok {
		t.Error("Data is not as expected")
	}
	if resp != nil {
		t.Log(resp.Status)
	}
}

func TestGetConfig(t *testing.T) {
	responseBody := "{\"Goodbye\":\"World\"}"
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(2 * time.Second)
		rw.Write([]byte(responseBody))
	}))
	defer server.Close()

	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	metaHttpClient := metahttp.NewClient(server.URL, logger, 1*time.Second)
	if metaHttpClient.GetConfig().Timeout != 1*time.Second {
		t.Error("Timeout is not matching")
	}
	if metaHttpClient.GetConfig().URL != server.URL {
		t.Error("URL is not matching")
	}
}

func TestMetaHTTPClientWithEmptyBasePath(t *testing.T) {
	responseBody := "{\"Goodbye\":\"World\"}"
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte(responseBody))
	}))
	defer server.Close()

	logLevel := &slog.LevelVar{} // INFO
	logLevel.Set(slog.LevelDebug)
	opts := &slog.HandlerOptions{
		AddSource: true,
		Level:     logLevel,
	}

	logger := slog.New(slog.NewJSONHandler(os.Stderr, opts))

	metaHttpClient := metahttp.NewClient("", logger, 10*time.Second)
	req := struct {
		Hello string
	}{
		Hello: "world",
	}
	var res struct {
		Goodbye string
	}

	ul := fmt.Sprintf("%s/test", server.URL)
	resp, err := metaHttpClient.Post(context.Background(), ul, map[string]string{}, req, &res)
	if err != nil {
		t.Error(err.Error())
	}
	if res.Goodbye != "World" {
		t.Error("Response body is not as expected")
	}
	if resp != nil {
		t.Log(resp.StatusCode)
	}
}

func TestMetaHTTPClientWithBadURLs(t *testing.T) {
	logLevel := &slog.LevelVar{} // INFO
	logLevel.Set(slog.LevelDebug)
	opts := &slog.HandlerOptions{
		AddSource: true,
		Level:     logLevel,
	}

	logger := slog.New(slog.NewJSONHandler(os.Stderr, opts))
	metaHttpClient := metahttp.NewClient("", logger, 10*time.Second)

	tcs := []struct {
		url string
	}{
		{
			url: "",
		},
		{
			url: "asdfsadf",
		},
		{
			url: "http::/google.com",
		},
		{
			url: "http//google.com",
		},
	}

	for _, tc := range tcs {
		var res map[string]any
		_, err := metaHttpClient.Get(context.Background(), tc.url, map[string]string{}, &res)
		if err == nil {
			t.Errorf("should be error but was passed to be a valid url, url: %s", tc.url)
		}

		if errors.Unwrap(err) != models.ErrBadURL {
			t.Errorf("not a bad url error, url: %s", tc.url)
		}
	}
}

func TestMetaHTTPClientUnAuthorised(t *testing.T) {
	responseBody := "{\"Goodbye\":\"World\"}"
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(401)
		rw.Write([]byte(responseBody))
	}))
	defer server.Close()

	logLevel := &slog.LevelVar{} // INFO
	logLevel.Set(slog.LevelDebug)
	opts := &slog.HandlerOptions{
		AddSource: true,
		Level:     logLevel,
	}

	logger := slog.New(slog.NewJSONHandler(os.Stderr, opts))

	metaHttpClient := metahttp.NewClient(server.URL, logger, 10*time.Second)
	req := struct {
		Hello string
	}{
		Hello: "world",
	}
	var res struct {
		Goodbye string
	}

	resp, err := metaHttpClient.Post(context.Background(), "/test", map[string]string{}, req, &res)
	if err == nil {
		t.Error("error should be present")
	}

	if resp == nil {
		t.Error("response should be present")
	}

	if resp != nil && resp.StatusCode != 401 {
		t.Error("invalid status code")
	}
}

func TestMetaHTTPClientInternalServerError(t *testing.T) {
	responseBody := "{\"error\":\"internal server error\"}"
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(500)
		rw.Write([]byte(responseBody))
	}))
	defer server.Close()

	logLevel := &slog.LevelVar{} // INFO
	logLevel.Set(slog.LevelDebug)
	opts := &slog.HandlerOptions{
		AddSource: true,
		Level:     logLevel,
	}

	logger := slog.New(slog.NewJSONHandler(os.Stderr, opts))

	metaHttpClient := metahttp.NewClient(server.URL, logger, 10*time.Second)
	req := struct {
		Hello string
	}{
		Hello: "world",
	}

	var res struct {
		Goodbye string
	}

	resp, err := metaHttpClient.Post(context.Background(), "/test", map[string]string{}, req, &res)
	if err == nil {
		t.Error("error should be present")
	}

	if resp == nil {
		t.Error("response should be present")
	}

	if resp != nil && resp.StatusCode != 500 {
		t.Error("invalid status code")
	}

	httpError, ok := err.(*models.HttpClientErrorResponse)
	if !ok {
		t.Error("error should be of type HttpClientErrorResponse")
	}

	if httpError.Err.Message != responseBody {
		t.Error("error message should be same as response body")
	}

	fmt.Println(httpError.Err.Message)
}
