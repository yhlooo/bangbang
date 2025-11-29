package commands

import (
	"fmt"

	"github.com/bombsimon/logrusr/v4"
	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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

// NewCommand 创建根命令
func NewCommand(name string) *cobra.Command {
	globalOpts := NewGlobalOptions()

	cmd := &cobra.Command{
		Use:           fmt.Sprintf("%s PIN", name),
		Short:         "Face-to-face group chat, file transfer.",
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
	}

	globalOpts.AddPFlags(cmd.PersistentFlags())

	cmd.AddCommand(
		newChatCommand(name),
		newScanCommand(),
	)

	return cmd
}
