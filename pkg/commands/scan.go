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
		Use:  "scan [ADDRESS]",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			scanAddr := "224.0.0.1:7134"
			if len(args) > 0 {
				scanAddr = args[0]
			}
			addr, err := net.ResolveUDPAddr("udp", scanAddr)
			if err != nil {
				return fmt.Errorf("resolve udp address %q error: %w", scanAddr, err)
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
