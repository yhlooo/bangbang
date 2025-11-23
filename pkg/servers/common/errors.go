package common

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	metav1 "github.com/yhlooo/bangbang/pkg/apis/meta/v1"
)

const (
	ErrReasonBadRequest          = "BadRequest"
	ErrReasonInternalServerError = "InternalServerError"
)

// HandleError 处理错误
func HandleError(ctx *gin.Context, err error) {
	status := StatusFromError(ctx, err)
	ctx.JSON(status.Code, status)
}

// StatusFromError 从 error 转为 *metav1.Status
func StatusFromError(ctx context.Context, err error) *metav1.Status {
	status := &metav1.Status{}
	if errors.As(err, &status) {
		return status
	}
	return NewInternalServerError(ctx, err.Error())
}

// NewError 创建错误
func NewError(ctx context.Context, code int, reason, message string) *metav1.Status {
	return &metav1.Status{
		APIMeta: metav1.NewAPIMeta(),
		Meta:    metav1.ObjectMeta{UID: RequestIDFromContext(ctx)},
		Code:    code,
		Reason:  reason,
		Message: message,
	}
}

// NewBadRequestError 创建 BadRequest 错误
func NewBadRequestError(ctx context.Context, message string) *metav1.Status {
	return NewError(ctx, http.StatusBadRequest, ErrReasonBadRequest, message)
}

// NewInternalServerError 创建 InternalServerError 错误
func NewInternalServerError(ctx context.Context, message string) *metav1.Status {
	return NewError(ctx, http.StatusInternalServerError, ErrReasonInternalServerError, message)
}
