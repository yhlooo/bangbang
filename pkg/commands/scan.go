package commands

import (
	"fmt"
	"io"
	"net"
	"os"

	"github.com/spf13/cobra"
)

// newScanCommand 创建 scan 子命令
func newScanCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "scan",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			addr, err := net.ResolveUDPAddr("udp", "224.0.0.1:2333")
			if err != nil {
				return fmt.Errorf("resolve udp add error: %w", err)
			}

			conn, err := net.ListenUDP("udp", addr)
			if err != nil {
				return fmt.Errorf("listen udp error: %w", err)
			}
			defer func() { _ = conn.Close() }()

			go func() {
				<-ctx.Done()
				_ = conn.Close()
			}()

			_, err = io.Copy(os.Stdout, conn)
			return err
		},
	}

	return cmd
}
