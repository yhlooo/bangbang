package common

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	metav1 "github.com/yhlooo/bangbang/pkg/apis/meta/v1"
)

const (
	ReasonOk                     = "Ok"
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

// NewStatus 创建 *metav1.Status
func NewStatus(ctx context.Context, code int, reason, message string) *metav1.Status {
	return &metav1.Status{
		APIMeta: metav1.NewAPIMeta(metav1.KindStatus),
		Meta:    metav1.ObjectMeta{UID: RequestIDFromContext(ctx)},
		Code:    code,
		Reason:  reason,
		Message: message,
	}
}

// NewOkStatus 创建正常状态
func NewOkStatus(ctx context.Context) *metav1.Status {
	return NewStatus(ctx, http.StatusOK, ReasonOk, "")
}

// NewBadRequestError 创建 BadRequest 错误
func NewBadRequestError(ctx context.Context, message string) *metav1.Status {
	return NewStatus(ctx, http.StatusBadRequest, ErrReasonBadRequest, message)
}

// NewInternalServerError 创建 InternalServerError 错误
func NewInternalServerError(ctx context.Context, message string) *metav1.Status {
	return NewStatus(ctx, http.StatusInternalServerError, ErrReasonInternalServerError, message)
}
