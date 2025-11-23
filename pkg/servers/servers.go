package servers

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"

	"github.com/yhlooo/bangbang/pkg/servers/chat"
	"github.com/yhlooo/bangbang/pkg/servers/common"
)

// Options 选项
type Options struct {
	ListenAddr string
}

// Complete 补全选项
func (o *Options) Complete() {
	if o.ListenAddr == "" {
		o.ListenAddr = ":0"
	}
}

// RunServer 运行服务
func RunServer(ctx context.Context, opts Options) (<-chan struct{}, error) {
	opts.Complete()

	logger := logr.FromContextOrDiscard(ctx)

	r := newGin(ctx)
	srv := &http.Server{
		Handler: r,
	}

	l, err := net.Listen("tcp", opts.ListenAddr)
	if err != nil {
		return nil, fmt.Errorf("listen on %q error: %w", opts.ListenAddr, err)
	}
	logger.Info(fmt.Sprintf("serve on %s", l.Addr().String()))

	done := make(chan struct{})
	go func() {
		defer close(done)
		if err := srv.Serve(l); err != nil && !errors.Is(err, context.Canceled) {
			logger.Error(err, "serve error")
		}
	}()
	go func() {
		<-ctx.Done()
		if err := srv.Shutdown(context.Background()); err != nil {
			logger.Error(err, "server shutdown error")
		}
	}()

	return done, nil
}

func newGin(reqCTX context.Context) *gin.Engine {
	r := gin.Default()
	r.ContextWithFallback = true

	r.Use(
		common.InjectRequestContext(reqCTX),
		common.InjectRequestID,
	)

	chatV1Group := r.Group("/chat/v1")

	chatServer := chat.NewServer()

	// 获取房间信息
	chatV1Group.GET("/rooms/:roomUID", typedHandler(chatServer.GetRoom))
	// 创建信道（加入房间）
	chatV1Group.POST("/rooms/:roomUID/channels", typedHandler(chatServer.CreateChannel))
	// 列出成员
	chatV1Group.GET("/rooms/:roomUID/members", typedHandler(chatServer.ListMembers))
	// 获取成员信息
	chatV1Group.GET("/rooms/:roomUID/members/:memberUID", typedHandler(chatServer.GetMember))

	return r
}

// typedHandler 有类型的 HTTP 处理器
func typedHandler[REQ any, RESP any](handler func(context.Context, *REQ) (*RESP, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := new(REQ)
		if err := ctx.ShouldBindUri(req); err != nil {
			common.HandleError(ctx, common.NewBadRequestError(ctx, fmt.Sprintf(
				"bind request uri error: %s",
				err.Error(),
			)))
			return
		}
		if err := ctx.ShouldBindQuery(req); err != nil {
			common.HandleError(ctx, common.NewBadRequestError(ctx, fmt.Sprintf(
				"bind request query error: %s",
				err.Error(),
			)))
			return
		}
		if withBody, ok := interface{}(req).(common.RequestWithBody); ok {
			if err := ctx.ShouldBindJSON(withBody.Body()); err != nil {
				common.HandleError(ctx, common.NewBadRequestError(ctx, fmt.Sprintf(
					"bind request body error: %s",
					err.Error(),
				)))
				return
			}
		}
		resp, err := handler(ctx, req)
		if err != nil {
			common.HandleError(ctx, err)
			return
		}
		ctx.JSON(200, resp)
	}
}
