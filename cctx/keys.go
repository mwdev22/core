package cctx

import "context"

type ContextKey string

const (
	RealIpKey ContextKey = "realIP"
	Limit     ContextKey = "limit"
	Offset    ContextKey = "offset"
)

func RealIP(ctx context.Context) string {
	if val := ctx.Value(RealIpKey); val != nil {
		return val.(string)
	}
	return ""
}
