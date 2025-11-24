package commands

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	"github.com/bombsimon/logrusr/v4"
	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/yhlooo/bangbang/pkg/chats/managers"
	"github.com/yhlooo/bangbang/pkg/servers"
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

	cmd := &cobra.Command{
		Use:           fmt.Sprintf("%s PIN", name),
		Short:         "Face-to-face group chat, file transfer.",
		Example:       exampleBuff.String(),
		Args:          cobra.ExactArgs(1),
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
				logrusLogger.Level = logrus.InfoLevel
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
			return run(cmd.Context())
		},
	}

	globalOpts.AddPFlags(cmd.PersistentFlags())

	return cmd
}

// run 运行
func run(ctx context.Context) error {
	done, err := servers.RunServer(ctx, servers.Options{
		ListenAddr:  ":0",
		ChatManager: managers.NewManager(),
	})
	if err != nil {
		return err
	}
	<-done
	return nil
}
