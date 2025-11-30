package commands

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	metav1 "github.com/yhlooo/bangbang/pkg/apis/meta/v1"
	"github.com/yhlooo/bangbang/pkg/chats/keys"
	"github.com/yhlooo/bangbang/pkg/managers"
	uitea "github.com/yhlooo/bangbang/pkg/ui/tty/tea"
)

// NewChatOptions 创建默认 ChatOptions
func NewChatOptions() ChatOptions {
	return ChatOptions{
		HTTPAddr:      ":0",
		DiscoveryAddr: "224.0.0.1:7134",
	}
}

// ChatOptions 选项
type ChatOptions struct {
	// 用户名
	Name string
	// HTTP 服务监听地址
	HTTPAddr string
	// 服务发现地址
	DiscoveryAddr string
}

// AddPFlags 将选项绑定到命令行参数
func (o *ChatOptions) AddPFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.Name, "name", "n", o.Name, "Your name")
	fs.StringVarP(&o.HTTPAddr, "listen", "l", o.HTTPAddr, "HTTP listen address")
	fs.StringVar(&o.DiscoveryAddr, "discovery-addr", o.DiscoveryAddr, "Transponder address")
}

var chatExampleTpl = template.Must(template.New("ChatCommand").
	Parse(`# Create or join a room using the specified PIN code. (e.g. 7134)
{{ .CommandName }} 7134
`))

func newChatCommand(parentName string) *cobra.Command {
	exampleBuff := &bytes.Buffer{}
	if err := chatExampleTpl.Execute(exampleBuff, map[string]interface{}{
		"CommandName": parentName + " chat",
	}); err != nil {
		panic(err)
	}

	opts := NewChatOptions()

	cmd := &cobra.Command{
		Use:     "chat PIN",
		Short:   "Start chat",
		Example: exampleBuff.String(),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runChat(cmd.Context(), opts, keys.HashKey(args[0]))
		},
	}

	opts.AddPFlags(cmd.Flags())

	return cmd
}

// run 运行
func runChat(ctx context.Context, opts ChatOptions, key keys.HashKey) error {
	selfUID := metav1.NewUID()

	mgr, err := managers.NewManager(managers.Options{
		Key:           key,
		OwnerUID:      selfUID,
		HTTPAddr:      opts.HTTPAddr,
		DiscoveryAddr: opts.DiscoveryAddr,
	})
	if err != nil {
		return fmt.Errorf("init manager error: %w", err)
	}

	// 运行服务
	if _, err := mgr.StartServer(ctx); err != nil {
		return fmt.Errorf("start server error: %w", err)
	}

	// 运行应答机
	if err := mgr.StartTransponder(ctx); err != nil {
		return fmt.Errorf("start transponder error: %w", err)
	}

	// 开始搜索上游
	if err := mgr.StartSearchUpstream(ctx); err != nil {
		return fmt.Errorf("start search upstream error: %w", err)
	}

	// 运行 UI
	ui := uitea.NewChatUI(mgr.SelfRoom(ctx), &metav1.ObjectMeta{
		UID:  selfUID,
		Name: opts.Name,
	})
	return ui.Run(ctx)
}
