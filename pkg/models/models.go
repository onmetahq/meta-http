package models

import "net/http"

type contextKey string

const (
	UserID           contextKey = "user-id"
	TenantID         contextKey = "tenant-id"
	RequestID        contextKey = "x-request-id"
	MerchantAPIKey   contextKey = "x-api-key"
	APIContextKey    contextKey = "apikey"
	AuthorizationKey contextKey = "Authorization"
)

var ContextKeys = []contextKey{UserID, TenantID, RequestID, MerchantAPIKey, APIContextKey, AuthorizationKey}

type ResponseData struct {
	Status     string // e.g. "200 OK"
	StatusCode int    // e.g. 200
	Header     http.Header
}
