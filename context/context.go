package context

import "context"

const (
	ReqIdKey  = "reqId"
	SessIdKey = "sessId"
)

// https://blog.gopheracademy.com/advent-2016/context-logging/

// WithRqId returns a context which knows its request ID
func WithReqId(ctx context.Context, reqId string) context.Context {
	return context.WithValue(ctx, ReqIdKey, reqId)
}

// WithSessionId returns a context which knows its session ID
func WithSessId(ctx context.Context, sessionId string) context.Context {
	return context.WithValue(ctx, SessIdKey, sessionId)
}
