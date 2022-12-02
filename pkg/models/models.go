package models

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
