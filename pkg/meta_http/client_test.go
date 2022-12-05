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
	resp, err := metaHttpClient.Post(context.Background(), "/test", map[string]string{}, req, &res)
	if err != nil {
		t.Error(err.Error())
	}
	if res.Goodbye != "World" {
		t.Error("Response body is not as expected")
	}
	t.Log(resp.Status)
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
	resp, err := metaHttpClient.Post(context.Background(), "/test", map[string]string{}, req, &res)
	if err == nil {
		t.Error("Supposed to fail with error")
	}
	t.Log(resp.Status)
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
	ctx = context.WithValue(ctx, models.UserID, "userId")
	ctx = context.WithValue(ctx, models.RequestID, "request-id")

	resp, err := metaHttpClient.Post(ctx, "/test", map[string]string{}, req, &res)
	if err != nil {
		t.Error(err.Error())
	}
	if res[string(models.UserID)] != "userId" {
		t.Error("Response body is not as expected")
	}
	t.Log(resp.Status)
}

func TestHeadersContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		res := map[string]string{}
		ctx := utils.FetchContextFromHeaders(context.Background(), req)
		if userId, ok := ctx.Value(models.UserID).(string); ok {
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
		string(models.UserID): "abcd",
	}

	resp, err := metaHttpClient.Post(ctx, "/test", headers, req, &res)
	if err != nil {
		t.Error(err.Error())
	}
	if res["data"] != "abcd" {
		t.Error("Response body is not as expected")
	}
	t.Log(resp.Status)
	res = map[string]string{}
	resp, err = metaHttpClient.Post(ctx, "/test", map[string]string{}, req, &res)
	if err != nil {
		t.Error(err.Error())
	}
	if _, ok := res["data"]; ok {
		t.Error("Data is not as expected")
	}
	t.Log(resp.Status)
}
