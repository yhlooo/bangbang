package tea

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-logr/logr"

	chatv1 "github.com/yhlooo/bangbang/pkg/apis/chat/v1"
	metav1 "github.com/yhlooo/bangbang/pkg/apis/meta/v1"
	"github.com/yhlooo/bangbang/pkg/chats/rooms"
)

// NewChatUI 创建聊天 UI
func NewChatUI(room rooms.Room, selfUID string) *ChatUI {
	return &ChatUI{
		selfUID: selfUID,
		room:    room,
	}
}

// ChatUI 聊天 UI
type ChatUI struct {
	ctx context.Context

	selfUID  string
	room     rooms.Room
	messages []*chatv1.Message

	vp    viewport.Model
	input textarea.Model
}

var _ tea.Model = (*ChatUI)(nil)

// Init 初始操作
func (ui *ChatUI) Init() tea.Cmd {
	return textarea.Blink
}

// Run 开始运行
func (ui *ChatUI) Run(ctx context.Context) error {
	ui.initInputBox()
	ui.vp = viewport.New(30, 5)
	ui.ctx = ctx

	msgCh, stop, err := ui.room.Listen(ctx)
	if err != nil {
		return fmt.Errorf("listen messages in room error: %w", err)
	}
	defer stop()

	p := tea.NewProgram(ui)

	go func() {
		for msg := range msgCh {
			p.Send(msg)
		}
	}()

	_, err = p.Run()
	return err
}

// Update 更新状态
func (ui *ChatUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	ctx := ui.ctx
	logger := logr.FromContextOrDiscard(ctx)

	var (
		inputCmd tea.Cmd
		vpCmd    tea.Cmd
	)

	ui.input, inputCmd = ui.input.Update(msg)
	ui.vp, vpCmd = ui.vp.Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		ui.vp.Width = msg.Width
		ui.input.SetWidth(msg.Width)
		ui.vp.Height = msg.Height - ui.input.Height() - 3

		ui.vp.SetContent(lipgloss.NewStyle().Width(ui.vp.Width).Render(ui.messagesContent()))
		ui.vp.GotoBottom()

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyCtrlD:
			fmt.Println(ui.input.Value())
			if err := ui.room.Close(ctx); err != nil {
				logger.Error(err, "close room error")
			}
			return ui, tea.Quit
		case tea.KeyEnter:
			err := ui.room.CreateMessage(ctx, &chatv1.Message{
				APIMeta: metav1.NewAPIMeta(chatv1.KindMessage),
				From: metav1.ObjectMeta{
					UID: ui.selfUID,
				},
				Content: chatv1.MessageContent{
					Text: &chatv1.TextMessageContent{Content: ui.input.Value()},
				},
			})
			if err != nil {
				logger.Error(err, "send message to room error")
			}
			ui.input.Reset()
			ui.vp.GotoBottom()
		}

	case *chatv1.Message:
		ui.messages = append(ui.messages, msg)
		ui.vp.SetContent(lipgloss.NewStyle().Width(ui.vp.Width).Render(ui.messagesContent()))

	case error:
		logger.Error(msg, "error")
		return ui, nil
	}

	return ui, tea.Batch(inputCmd, vpCmd)
}

// View 生成显示内容
func (ui *ChatUI) View() string {
	return ui.vp.View() + "\n\n" + ui.input.View()
}

// initInputBox 初始化输入框
func (ui *ChatUI) initInputBox() {
	ui.input = textarea.New()
	ui.input.Placeholder = "Send a message..."
	ui.input.Focus()
	ui.input.Prompt = "┃ "
	ui.input.CharLimit = 1024
	ui.input.SetWidth(30)
	ui.input.SetHeight(3)
	ui.input.FocusedStyle.CursorLine = lipgloss.NewStyle() // Remove cursor line styling
	ui.input.ShowLineNumbers = false
	ui.input.KeyMap.InsertNewline.SetEnabled(false)
}

// messagesContent 获取消息文本形式展示的内容
func (ui *ChatUI) messagesContent() string {
	retLines := make([]string, 0, len(ui.messages)*2)
	for _, msg := range ui.messages {
		if msg.Content.Text != nil {
			retLines = append(retLines, msg.From.UID, msg.Content.Text.Content, "")
		}
	}
	return strings.Join(retLines, "\n")
}
