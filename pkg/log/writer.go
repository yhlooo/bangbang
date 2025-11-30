package log

import (
	"context"
	"io"
)

type logWriterContextKey struct{}

// WriterFromContext 从上下文获取输出日志的 io.Writer
func WriterFromContext(ctx context.Context) io.Writer {
	w, ok := ctx.Value(logWriterContextKey{}).(io.Writer)
	if !ok {
		return io.Discard
	}
	return w
}

// ContextWithWriter 返回包含指定 io.Writer 的 context.Context
func ContextWithWriter(parent context.Context, w io.Writer) context.Context {
	return context.WithValue(parent, logWriterContextKey{}, w)
}
