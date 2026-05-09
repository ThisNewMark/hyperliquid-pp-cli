// Hand-rewritten: typed usdSend + signed user-signed POST.

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

func newExchangeSendUsdCmd(flags *rootFlags) *cobra.Command {
	var destination string
	var amount string

	cmd := &cobra.Command{
		Use:   "send-usd",
		Short: "Transfer USDC on the perps account (signed user action — MAIN wallet)",
		Annotations: map[string]string{
			"pp:endpoint": "exchange.send-usd",
			"pp:method":   "POST",
			"pp:path": "/exchange/sendUsd", "mcp:hidden": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, name := range []string{"destination", "amount"} {
				if !cmd.Flags().Changed(name) && !flags.dryRun {
					return fmt.Errorf("required flag --%s not set", name)
				}
			}
			dest := strings.ToLower(strings.TrimSpace(destination))
			timeMs := uint64(time.Now().UnixMilli())
			action := map[string]any{
				"type":        "usdSend",
				"destination": dest,
				"amount":      amount,
				"time":        timeMs,
			}
			primary, payloadTypes := sign.UsdSendSpec()

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
	cmd.Flags().StringVar(&destination, "destination", "", "Recipient 0x address (required)")
	cmd.Flags().StringVar(&amount, "amount", "", "USDC amount as a string, e.g. \"10\" (required)")
	return cmd
}
