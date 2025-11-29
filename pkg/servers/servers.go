package servers

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"

	"github.com/yhlooo/bangbang/pkg/chats/rooms"
	"github.com/yhlooo/bangbang/pkg/servers/chat"
	"github.com/yhlooo/bangbang/pkg/servers/common"
)

// Options 选项
type Options struct {
	ListenAddr string
	Room       rooms.Room
}

// Complete 补全选项
func (o *Options) Complete() {
	if o.ListenAddr == "" {
		o.ListenAddr = ":0"
	}
}

// RunServer 运行服务
func RunServer(ctx context.Context, opts Options) (net.Addr, <-chan struct{}, error) {
	opts.Complete()

	logger := logr.FromContextOrDiscard(ctx)

	if !logger.V(1).Enabled() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := newGin(ctx, opts.Room)
	srv := &http.Server{
		Handler: r,
	}

	l, err := net.Listen("tcp", opts.ListenAddr)
	if err != nil {
		return nil, nil, fmt.Errorf("listen on %q error: %w", opts.ListenAddr, err)
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

	return l.Addr(), done, nil
}

func newGin(reqCTX context.Context, room rooms.Room) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.ContextWithFallback = true

	r.Use(
		common.InjectRequestContext(reqCTX),
		common.InjectRequestID,
	)

	chatV1Group := r.Group("/chat/v1")

	chatServer := chat.NewServer(room)

	chatV1Group.GET("/info", typedHandler(chatServer.GetInfo))
	// 创建消息（发送消息）
	chatV1Group.POST("/messages", typedHandler(chatServer.CreateMessage))
	// 监听消息
	chatV1Group.GET("/messages", typedHandler(chatServer.ListenMessages))

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
