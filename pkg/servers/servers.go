package servers

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	stdlog "log"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"

	"github.com/yhlooo/bangbang/pkg/chats/rooms"
	"github.com/yhlooo/bangbang/pkg/log"
	"github.com/yhlooo/bangbang/pkg/servers/chat"
	"github.com/yhlooo/bangbang/pkg/servers/common"
	"github.com/yhlooo/bangbang/pkg/signatures"
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
func RunServer(ctx context.Context, opts Options) (net.Addr, string, <-chan struct{}, error) {
	opts.Complete()

	logger := logr.FromContextOrDiscard(ctx)

	if !logger.V(1).Enabled() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := newGin(ctx, opts.Room, log.WriterFromContext(ctx))
	srv := &http.Server{
		Handler:  r,
		ErrorLog: stdlog.New(log.WriterFromContext(ctx), "", stdlog.LstdFlags),
	}

	// 生成ECC自签名证书
	certPEM, keyPEM, err := GenerateECCCertificate("bangbang")
	if err != nil {
		return nil, "", nil, fmt.Errorf("generate certificate error: %w", err)
	}

	// 创建证书对
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, "", nil, fmt.Errorf("create certificate pair error: %w", err)
	}

	// 配置TLS
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	// 监听
	l, err := tls.Listen("tcp", opts.ListenAddr, tlsConfig)
	if err != nil {
		return nil, "", nil, fmt.Errorf("listen on %q error: %w", opts.ListenAddr, err)
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

	return l.Addr(), signatures.SignCert(cert.Leaf.Raw), done, nil
}

func newGin(reqCTX context.Context, room rooms.Room, logWriter io.Writer) *gin.Engine {
	gin.DefaultWriter = logWriter
	gin.DefaultErrorWriter = logWriter
	r := gin.New()
	r.ContextWithFallback = true

	r.Use(
		gin.Recovery(),
		gin.LoggerWithWriter(logWriter),
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
