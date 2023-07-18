package models

import (
	"fmt"
	"net/http"
	"time"
)

type contextKey string

const (
	UserID           contextKey = "user-id"
	TenantID         contextKey = "tenant-id"
	RequestID        contextKey = "x-request-id"
	MerchantAPIKey   contextKey = "x-api-key"
	APIContextKey    contextKey = "apikey"
	AuthorizationKey contextKey = "Authorization"
	XForwardedFor    contextKey = "X-Forwarded-For"
)

var ContextKeys = []contextKey{UserID, TenantID, RequestID, MerchantAPIKey, APIContextKey, AuthorizationKey, XForwardedFor}

type ResponseData struct {
	Status     string // e.g. "200 OK"
	StatusCode int    // e.g. 200
	Header     http.Header
}

type Retry struct {
	MaxRetries        int
	DelayBetweenRetry time.Duration
	Validator         func(int) bool
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
