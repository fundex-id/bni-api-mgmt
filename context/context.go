package context

import "context"

const (
	HttpReqIdKey  = "httpReqId"
	HttpSessIdKey = "httpSessId"
	BniSessIdKey  = "bniSessId"
)

// https://blog.gopheracademy.com/advent-2016/context-logging/

// WithRqId returns a context which knows its HTTP request ID
func WithHttpReqId(ctx context.Context, reqId string) context.Context {
	return context.WithValue(ctx, HttpReqIdKey, reqId)
}

// WithSessionId returns a context which knows its HTTP session ID
func WithHttpSessId(ctx context.Context, sessionId string) context.Context {
	return context.WithValue(ctx, HttpSessIdKey, sessionId)
}

// WithBNISessionId returns a context which knows its BNI session ID
func WithBniSessId(ctx context.Context, sessionId string) context.Context {
	return context.WithValue(ctx, BniSessIdKey, sessionId)
}
