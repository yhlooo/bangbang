package commands

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/bombsimon/logrusr/v4"
	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/yhlooo/bangbang/pkg/log"
	"github.com/yhlooo/bangbang/pkg/version"
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
	// 是否开启调试模式
	Debug bool
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
	fs.BoolVar(&o.Debug, "debug", false, "Run in debug mode")
}

// NewCommand 创建根命令
func NewCommand(name string) *cobra.Command {
	globalOpts := NewGlobalOptions()

	cmd := &cobra.Command{
		Use:           fmt.Sprintf("%s PIN", name),
		Short:         "Face-to-face group chat, file transfer.",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version.Version,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if err := globalOpts.Validate(); err != nil {
				return err
			}

			logWriter := io.Writer(os.Stderr)
			switch cmd.Name() {
			case "chat":
				logWriter = io.Discard
				if globalOpts.Debug || globalOpts.Verbosity >= 1 {
					logPath := filepath.Join(os.ExpandEnv("$HOME"), ".bangbang", "bang.log")
					if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
						return fmt.Errorf("create log directory %q error: %w", filepath.Dir(logPath), err)
					}
					f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
					if err != nil {
						return fmt.Errorf("create log file %q error: %w", logPath, err)
					}
					logWriter = f
				}
			default:
			}
			ctx = log.ContextWithWriter(ctx, logWriter)

			// 初始化 logger
			logrusLogger := logrus.New()
			logrusLogger.SetOutput(logWriter)
			switch globalOpts.Verbosity {
			case 0:
				logrusLogger.Level = logrus.InfoLevel
			case 1:
				logrusLogger.Level = logrus.DebugLevel
			default:
				logrusLogger.Level = logrus.TraceLevel
			}
			logger := logrusr.New(logrusLogger)
			cmd.SetContext(logr.NewContext(ctx, logger))

			return nil
		},
	}

	globalOpts.AddPFlags(cmd.PersistentFlags())

	cmd.AddCommand(
		newChatCommand(name),
		newScanCommand(),
		newVersionCommand(),
	)

	return cmd
}
