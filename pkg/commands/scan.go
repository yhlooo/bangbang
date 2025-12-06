package commands

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/yhlooo/bangbang/pkg/discovery"
	"github.com/yhlooo/bangbang/pkg/signatures"
)

// NewScanOptions 创建默认 ScanOptions
func NewScanOptions() ScanOptions {
	return ScanOptions{
		Watch:             false,
		Key:               "",
		Duration:          3 * time.Second,
		RequestInterval:   time.Second,
		CheckAvailability: true,
	}
}

// ScanOptions scan 子命令选项
type ScanOptions struct {
	Watch             bool
	Key               string
	Duration          time.Duration
	RequestInterval   time.Duration
	CheckAvailability bool
}

// AddPFlags 将选项绑定到命令行参数
func (o *ScanOptions) AddPFlags(fs *pflag.FlagSet) {
	fs.BoolVarP(&o.Watch, "watch", "w", o.Watch, "Keep watching")
	fs.StringVarP(&o.Key, "key", "k", o.Key, "Room Key")
	fs.DurationVarP(&o.Duration, "duration", "d", o.Duration, "Scan duration")
	fs.DurationVar(&o.RequestInterval, "interval", o.RequestInterval, "Send scan request interval")
	fs.BoolVar(&o.CheckAvailability, "check", o.CheckAvailability, "Check room endpoints availability")
}

// newScanCommand 创建 scan 子命令
func newScanCommand() *cobra.Command {
	opts := NewScanOptions()

	cmd := &cobra.Command{
		Use:  "scan [ADDRESS]",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			addr := "224.0.0.1:7134"
			if len(args) > 0 {
				addr = args[0]
			}

			d := discovery.NewUDPDiscoverer(addr)
			var key signatures.Key
			if opts.Key != "" {
				key = signatures.Key(opts.Key)
			}

			ticker := time.NewTicker(opts.Duration + time.Second)
			for {
				ret, err := d.Search(ctx, key, discovery.SearchOptions{
					Duration:          opts.Duration,
					RequestInterval:   opts.RequestInterval,
					CheckAvailability: opts.CheckAvailability,
				})
				if err != nil {
					return fmt.Errorf("search rooms error: %w", err)
				}
				showScanResult(ret)
				if !opts.Watch {
					break
				}
				select {
				case <-ctx.Done():
					return nil
				case <-ticker.C:
				}
			}

			return nil
		},
	}

	opts.AddPFlags(cmd.Flags())

	return cmd
}

// showScanResult 展示搜索结果
func showScanResult(result []discovery.Room) {
	for _, room := range result {
		fmt.Printf("      UID : %s\n", room.Info.UID)
		if ownerUID := room.Info.Owner.UID; !ownerUID.IsNil() {
			fmt.Printf("    Owner : %s\n", ownerUID)
		}
		if room.Info.Signature != "" {
			fmt.Printf("Signature : %s\n", room.Info.Signature)
		}
		if room.Info.CertSign != "" {
			fmt.Printf("Cert Sign : %s\n", room.Info.CertSign)
		}
		if len(room.Info.Endpoints) > 0 {
			fmt.Println("Endpoints :")
			for _, endpoint := range room.Info.Endpoints {
				if endpoint == room.AvailableEndpoint {
					fmt.Printf("            %s (Available)\n", endpoint)
				} else {
					fmt.Printf("            %s\n", endpoint)
				}
			}
		} else {
			fmt.Println("Endpoints : []")
		}
		fmt.Println("---")
	}
}
