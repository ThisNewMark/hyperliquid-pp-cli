// Hand-rewritten: typed ScheduleCancelAction + signed L1 POST.

package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"hyperliquid-pp-cli/internal/client"
	"hyperliquid-pp-cli/internal/sign"
)

func newExchangeScheduleCancelCmd(flags *rootFlags) *cobra.Command {
	var timeMs int64
	var clear bool
	var walletFlag string

	cmd := &cobra.Command{
		Use:   "schedule-cancel",
		Short: "Dead-man-switch: schedule a cancel-all at a future timestamp (signed L1 action)",
		Long: `Sets a server-side timer that cancels every open order at --time (ms epoch).
Must be at least 5 seconds in the future. Pass --clear to remove the timer.

Hyperliquid limits: max 10 trigger sets per day, reset at 00:00 UTC.`,
		Annotations: map[string]string{
			"pp:endpoint": "exchange.schedule-cancel",
			"pp:method":   "POST",
			"pp:path": "/exchange/scheduleCancel", "mcp:hidden": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			var t *int64
			if !clear {
				if !cmd.Flags().Changed("time") && !flags.dryRun {
					return fmt.Errorf("either --time <ms-epoch> or --clear is required")
				}
				v := timeMs
				t = &v
			}
			action := sign.NewScheduleCancelAction(t)

			c, err := flags.newClient()
			if err != nil {
				return err
			}
			cfg := c.Config
			kind := walletAuto
			if strings.EqualFold(walletFlag, "main") {
				kind = walletMain
			} else if strings.EqualFold(walletFlag, "agent") {
				kind = walletAgent
			}
			key, signerAddr, err := walletKey(cfg, kind)
			if err != nil {
				if !flags.dryRun {
					return err
				}
				signerAddr = "<no key — dry-run>"
			}
			opts := client.SignOpts{IsMainnet: cfg.IsMainnet}
			if flags.dryRun {
				return printSignedDryRun(cmd, action, signerAddr, opts, nil, key)
			}
			data, status, err := c.PostSignedL1(action, key, opts)
			if err != nil {
				return classifyAPIError(err, flags)
			}
			return printSignedResult(cmd, flags, "/exchange", data, status, action, signerAddr, nil)
		},
	}
	cmd.Flags().Int64Var(&timeMs, "time", 0, "Trigger time in ms epoch (≥5s in the future)")
	cmd.Flags().BoolVar(&clear, "clear", false, "Clear the scheduled cancel instead of setting one")
	cmd.Flags().StringVar(&walletFlag, "wallet", "", "Override active wallet")
	return cmd
}
