// Hand-rewritten: typed withdraw3 + signed user-signed POST.

package cli

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"hyperliquid-pp-cli/internal/client"
	"hyperliquid-pp-cli/internal/sign"
)

func newExchangeWithdrawCmd(flags *rootFlags) *cobra.Command {
	var destination string
	var amount string

	cmd := &cobra.Command{
		Use:   "withdraw",
		Short: "Withdraw USDC from Hyperliquid to L1 Arbitrum (~5min, $1 fee, MAIN wallet)",
		Long:  "Issues a withdraw3 action. Hyperliquid takes ~5 minutes to bridge USDC to your destination. A $1 fee is deducted server-side.",
		Annotations: map[string]string{
			"pp:endpoint": "exchange.withdraw",
			"pp:method":   "POST",
			"pp:path": "/exchange/withdraw", "mcp:hidden": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, name := range []string{"destination", "amount"} {
				if !cmd.Flags().Changed(name) && !flags.dryRun {
					return fmt.Errorf("required flag --%s not set", name)
				}
			}
			dest := strings.ToLower(strings.TrimSpace(destination))
			timeMs := uint64(time.Now().UnixMilli())
			action := map[string]any{
				"type":        "withdraw3",
				"destination": dest,
				"amount":      amount,
				"time":        timeMs,
			}
			primary, payloadTypes := sign.WithdrawSpec()

			c, err := flags.newClient()
			if err != nil {
				return err
			}
			cfg := c.Config
			if flags.dryRun {
				return printApproveDryRun(cmd, action, primary, payloadTypes)
			}
			mainKey, mainAddr, err := walletKey(cfg, walletMain)
			if err != nil {
				return err
			}
			data, status, err := c.PostSignedUser(action, primary, payloadTypes, mainKey, client.SignOpts{IsMainnet: cfg.IsMainnet, Nonce: timeMs})
			if err != nil {
				return classifyAPIError(err, flags)
			}
			result := map[string]any{"destination": dest, "amount": amount, "signer": mainAddr, "status": status, "success": status >= 200 && status < 300}
			var parsed any
			if err := json.Unmarshal(data, &parsed); err == nil {
				result["data"] = parsed
			}
			enc, _ := json.MarshalIndent(result, "", "  ")
			fmt.Fprintln(cmd.OutOrStdout(), string(enc))
			return nil
		},
	}
	cmd.Flags().StringVar(&destination, "destination", "", "L1 Arbitrum recipient 0x address (required)")
	cmd.Flags().StringVar(&amount, "amount", "", "USDC amount as a string, e.g. \"50\" (required)")
	return cmd
}
