package metahttp_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-kit/log"
	metahttp "github.com/krishnateja262/meta-http/meta_http"
)

func TestMetaHTTPClient(t *testing.T) {
	responseBody := "{\"Goodbye\":\"World\"}"
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte(responseBody))
	}))
	defer server.Close()

	logger := log.NewJSONLogger(os.Stderr)
	logger = log.NewSyncLogger(logger)

	metaHttpClient := metahttp.NewClient(server.URL, logger, 10*time.Second)
	req := struct {
		Hello string
	}{
		Hello: "world",
	}
	var res struct {
		Goodbye string
	}
	err := metaHttpClient.Post(context.Background(), "/test", map[string]string{}, req, &res)
	if err != nil {
		t.Error(err.Error())
	}
	if res.Goodbye != "World" {
		t.Error("Response body is not as expected")
	}
}

func TestTimeoutScenario(t *testing.T) {
	responseBody := "{\"Goodbye\":\"World\"}"
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(2 * time.Second)
		rw.Write([]byte(responseBody))
	}))
	defer server.Close()

	logger := log.NewJSONLogger(os.Stderr)
	logger = log.NewSyncLogger(logger)

	metaHttpClient := metahttp.NewClient(server.URL, logger, 1*time.Second)
	req := struct {
		Hello string
	}{
		Hello: "world",
	}
	var res struct {
		Goodbye string
	}
	err := metaHttpClient.Post(context.Background(), "/test", map[string]string{}, req, &res)
	if err == nil {
		t.Error("Supposed to fail with error")
	}
}

func TestContextHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		userId := req.Header.Get(string(metahttp.UserID))
		res := map[string]string{
			string(metahttp.UserID): userId,
		}
		bytes, _ := json.Marshal(res)
		rw.Write(bytes)
	}))
	defer server.Close()

	logger := log.NewJSONLogger(os.Stderr)
	logger = log.NewSyncLogger(logger)

	metaHttpClient := metahttp.NewClient(server.URL, logger, 1*time.Second)
	req := struct {
		Hello string
	}{
		Hello: "world",
	}
	var res map[string]string
	ctx := context.Background()
	ctx = context.WithValue(ctx, metahttp.UserID, "userId")
	ctx = context.WithValue(ctx, metahttp.RequestID, "request-id")

	err := metaHttpClient.Post(ctx, "/test", map[string]string{}, req, &res)
	if err != nil {
		t.Error(err.Error())
	}
	if res[string(metahttp.UserID)] != "userId" {
		t.Error("Response body is not as expected")
	}
}

func TestHeadersContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		res := map[string]string{}
		ctx := metahttp.FetchContextFromHeaders(context.Background(), req)
		if userId, ok := ctx.Value(metahttp.UserID).(string); ok {
			res["data"] = userId
		}
		bytes, _ := json.Marshal(res)
		rw.Write(bytes)
	}))
	defer server.Close()

	logger := log.NewJSONLogger(os.Stderr)
	logger = log.NewSyncLogger(logger)

	metaHttpClient := metahttp.NewClient(server.URL, logger, 1*time.Second)
	req := struct {
		Hello string
	}{
		Hello: "world",
	}
	var res map[string]string
	ctx := context.Background()

	headers := map[string]string{
		string(metahttp.UserID): "abcd",
	}

	err := metaHttpClient.Post(ctx, "/test", headers, req, &res)
	if err != nil {
		t.Error(err.Error())
	}
	if res["data"] != "abcd" {
		t.Error("Response body is not as expected")
	}

	res = map[string]string{}
	err = metaHttpClient.Post(ctx, "/test", map[string]string{}, req, &res)
	if err != nil {
		t.Error(err.Error())
	}
	if _, ok := res["data"]; ok {
		t.Error("Data is not as expected")
	}
}
