// Hand-rewritten: typed spotSend + signed user-signed POST.

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

func newExchangeSendSpotCmd(flags *rootFlags) *cobra.Command {
	var destination string
	var token string
	var amount string

	cmd := &cobra.Command{
		Use:   "send-spot",
		Short: "Transfer a spot asset to another address (signed user action — MAIN wallet)",
		Annotations: map[string]string{
			"pp:endpoint": "exchange.send-spot",
			"pp:method":   "POST",
			"pp:path": "/exchange/sendSpot", "mcp:hidden": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, name := range []string{"destination", "token", "amount"} {
				if !cmd.Flags().Changed(name) && !flags.dryRun {
					return fmt.Errorf("required flag --%s not set", name)
				}
			}
			dest := strings.ToLower(strings.TrimSpace(destination))
			// time field carries the millis timestamp for spotSend (acts as nonce).
			timeMs := uint64(time.Now().UnixMilli())
			action := map[string]any{
				"type":        "spotSend",
				"destination": dest,
				"token":       token,
				"amount":      amount,
				"time":        timeMs,
			}
			primary, payloadTypes := sign.SpotSendSpec()

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
			result := map[string]any{"destination": dest, "token": token, "amount": amount, "signer": mainAddr, "status": status, "success": status >= 200 && status < 300}
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
	cmd.Flags().StringVar(&token, "token", "", "Token in 'name:tokenId' format, e.g. 'PURR:0xc4bf...' (required)")
	cmd.Flags().StringVar(&amount, "amount", "", "Amount as a string (required)")
	return cmd
}
