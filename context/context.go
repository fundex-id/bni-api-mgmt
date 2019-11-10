package context

import "context"

const (
	HTTPReqIDKey  = "httpReqID"
	HTTPSessIDKey = "httpSessID"
	BNISessIDKey  = "bniSessID"
)

// https://blog.gopheracademy.com/advent-2016/context-logging/

// WithRqId returns a context which knows its HTTP request ID
func WithHTTPReqID(ctx context.Context, reqId string) context.Context {
	return context.WithValue(ctx, HTTPReqIDKey, reqId)
}

// WithSessionId returns a context which knows its HTTP session ID
func WithHTTPSessID(ctx context.Context, sessionId string) context.Context {
	return context.WithValue(ctx, HTTPSessIDKey, sessionId)
}

// WithBNISessionId returns a context which knows its BNI session ID
func WithBNISessID(ctx context.Context, sessionId string) context.Context {
	return context.WithValue(ctx, BNISessIDKey, sessionId)
}
