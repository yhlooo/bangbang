package common

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	"github.com/google/uuid"
)

// InjectRequestContext 注入请求上下文
func InjectRequestContext(reqCTX context.Context) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Request = ctx.Request.WithContext(reqCTX)
	}
}

const RequestIDHeader = "X-Request-Id"

// InjectRequestID 注入请求 ID
func InjectRequestID(ctx *gin.Context) {
	reqID := uuid.New().String()
	// 注入到上下文
	reqCTX := ctx.Request.Context()
	reqCTX = NewContextWithRequestID(reqCTX, reqID)
	// 注入到 logger
	reqCTX = logr.NewContext(reqCTX, logr.FromContextOrDiscard(ctx).WithValues("requestID", reqID))
	ctx.Request = ctx.Request.WithContext(reqCTX)
	// 注入到响应头
	ctx.Header(RequestIDHeader, reqID)
}

type reqIDContextKey struct{}

// RequestIDFromContext 从 ctx 获取请求 ID
func RequestIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(reqIDContextKey{}).(string)
	return v
}

// NewContextWithRequestID 返回携带请求 ID 的上下文
func NewContextWithRequestID(parent context.Context, requestID string) context.Context {
	return context.WithValue(parent, reqIDContextKey{}, requestID)
}
