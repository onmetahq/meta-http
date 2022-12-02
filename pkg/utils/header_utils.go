package utils

import (
	"context"
	"net/http"

	"github.com/krishnateja262/meta-http/models"
)

func FetchHeadersFromContext(ctx context.Context) map[string]string {
	ctxHeaders := map[string]string{}
	for _, key := range models.ContextKeys {
		val, ok := ctx.Value(key).(string)
		if ok {
			ctxHeaders[string(key)] = val
		}
	}

	return ctxHeaders
}

func FetchContextFromHeaders(ctx context.Context, r *http.Request) context.Context {
	for _, key := range models.ContextKeys {
		val := r.Header.Get(string(key))
		if val != "" {
			ctx = context.WithValue(ctx, key, val)
		}
	}
	return ctx
}
