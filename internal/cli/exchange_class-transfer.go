// Hand-rewritten: typed usdClassTransfer + signed user-signed POST.

package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"hyperliquid-pp-cli/internal/client"
	"hyperliquid-pp-cli/internal/sign"
)

func newExchangeClassTransferCmd(flags *rootFlags) *cobra.Command {
	var amount string
	var toPerp bool

	cmd := &cobra.Command{
		Use:   "class-transfer",
		Short: "Move USDC between perp and spot accounts (signed user action — MAIN wallet)",
		Annotations: map[string]string{
			"pp:endpoint": "exchange.class-transfer",
			"pp:method":   "POST",
			"pp:path": "/exchange/classTransfer", "mcp:hidden": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cmd.Flags().Changed("amount") && !flags.dryRun {
				return fmt.Errorf("required flag --amount not set")
			}
			action := map[string]any{
				"type":   "usdClassTransfer",
				"amount": amount,
				"toPerp": toPerp,
			}
			primary, payloadTypes := sign.UsdClassTransferSpec()

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
			data, status, err := c.PostSignedUser(action, primary, payloadTypes, mainKey, client.SignOpts{IsMainnet: cfg.IsMainnet})
			if err != nil {
				return classifyAPIError(err, flags)
			}
			result := map[string]any{"amount": amount, "to_perp": toPerp, "signer": mainAddr, "status": status, "success": status >= 200 && status < 300}
			var parsed any
			if err := json.Unmarshal(data, &parsed); err == nil {
				result["data"] = parsed
			}
			enc, _ := json.MarshalIndent(result, "", "  ")
			fmt.Fprintln(cmd.OutOrStdout(), string(enc))
			return nil
		},
	}
	cmd.Flags().StringVar(&amount, "amount", "", "USDC amount as a string, e.g. \"10.5\"")
	cmd.Flags().BoolVar(&toPerp, "to-perp", true, "true = spot → perp; false = perp → spot")
	return cmd
}
