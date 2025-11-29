package commands

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	"github.com/bombsimon/logrusr/v4"
	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/yhlooo/bangbang/pkg/chats/managers"
	"github.com/yhlooo/bangbang/pkg/chats/rooms"
	"github.com/yhlooo/bangbang/pkg/servers"
	uitea "github.com/yhlooo/bangbang/pkg/ui/tty/tea"
)

// NewGlobalOptions 创建一个默认 GlobalOptions
func NewGlobalOptions() GlobalOptions {
	return GlobalOptions{
		Verbosity: 0,
	}
}

// GlobalOptions 全局选项
type GlobalOptions struct {
	// 日志数量级别（ 0 / 1 / 2 ）
	Verbosity uint32
}

// Validate 校验选项是否合法
func (o *GlobalOptions) Validate() error {
	if o.Verbosity > 2 {
		return fmt.Errorf("invalid log verbosity: %d (expected: 0, 1 or 2)", o.Verbosity)
	}
	return nil
}

// AddPFlags 将选项绑定到命令行参数
func (o *GlobalOptions) AddPFlags(fs *pflag.FlagSet) {
	fs.Uint32VarP(&o.Verbosity, "verbose", "v", o.Verbosity, "Number for the log level verbosity (0, 1, or 2)")
}

func NewOptions() Options {
	return Options{
		HostMode:  false,
		GuestMode: false,
		Address:   ":2333",
	}
}

// Options 选项
type Options struct {
	HostMode  bool
	GuestMode bool
	Address   string
}

// AddPFlags 将选项绑定到命令行参数
func (o *Options) AddPFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&o.HostMode, "host", o.HostMode, "Run in host mode")
	fs.BoolVar(&o.GuestMode, "guest", o.GuestMode, "Run in guest mode")
	fs.StringVar(&o.Address, "addr", o.Address, "Listen or connect address")
}

var rootExampleTpl = template.Must(template.New("RootCommand").
	Parse(`# Create or join a room using the specified PIN code. (e.g. 7134)
{{ .CommandName }} 7134
`))

// NewCommand 创建根命令
func NewCommand(name string) *cobra.Command {
	exampleBuff := &bytes.Buffer{}
	if err := rootExampleTpl.Execute(exampleBuff, map[string]interface{}{
		"CommandName": name,
	}); err != nil {
		panic(err)
	}

	globalOpts := NewGlobalOptions()
	opts := NewOptions()

	cmd := &cobra.Command{
		Use:           fmt.Sprintf("%s PIN", name),
		Short:         "Face-to-face group chat, file transfer.",
		Example:       exampleBuff.String(),
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := globalOpts.Validate(); err != nil {
				return err
			}

			// 初始化 logger
			logrusLogger := logrus.New()
			switch globalOpts.Verbosity {
			case 0:
				logrusLogger.Level = logrus.ErrorLevel
			case 1:
				logrusLogger.Level = logrus.DebugLevel
			default:
				logrusLogger.Level = logrus.TraceLevel
			}
			logger := logrusr.New(logrusLogger)
			cmd.SetContext(logr.NewContext(cmd.Context(), logger))

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), opts)
		},
	}

	globalOpts.AddPFlags(cmd.PersistentFlags())
	opts.AddPFlags(cmd.Flags())

	return cmd
}

// run 运行
func run(ctx context.Context, opts Options) error {
	logger := logr.FromContextOrDiscard(ctx)
	if opts.GuestMode {
		logger.Info("run in guest mode")
		return runInGuestMode(ctx, opts)
	}
	logger.Info("run in host mode")
	return runInHostMode(ctx, opts)
}

// runInHostMode 以房主模式运行
func runInHostMode(ctx context.Context, opts Options) error {
	mgr := managers.NewManager()
	room, err := mgr.GetLocalRoom(ctx, managers.DefaultRoomID)
	if err != nil {
		return fmt.Errorf("get room error: %w", err)
	}
	_, err = servers.RunServer(ctx, servers.Options{
		ListenAddr:  opts.Address,
		ChatManager: mgr,
	})
	if err != nil {
		return fmt.Errorf("run server error: %w", err)
	}

	ui := uitea.NewChatUI(room, uuid.New().String())
	return ui.Run(ctx)
}

// runInGuestMode 以客人模式运行
func runInGuestMode(ctx context.Context, opts Options) error {
	room := rooms.NewRemoteRoom("http://"+opts.Address, managers.DefaultRoomID)

	ui := uitea.NewChatUI(room, uuid.New().String())
	return ui.Run(ctx)
}
