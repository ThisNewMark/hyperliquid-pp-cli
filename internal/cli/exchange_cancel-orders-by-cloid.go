// Hand-rewritten: typed CancelCloidAction + signed L1 POST.

package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"hyperliquid-pp-cli/internal/client"
	"hyperliquid-pp-cli/internal/sign"
)

func newExchangeCancelOrdersByCloidCmd(flags *rootFlags) *cobra.Command {
	var bodyCancelsJSON string
	var walletFlag string

	cmd := &cobra.Command{
		Use:     "cancel-orders-by-cloid",
		Short:   "Cancel orders by client order id (signed L1 action)",
		Example: "  hyperliquid exchange cancel-orders-by-cloid --cancels '[{\"asset\":0,\"cloid\":\"0xabc...\"}]'",
		Annotations: map[string]string{
			"pp:endpoint": "exchange.cancel-orders-by-cloid",
			"pp:method":   "POST",
			"pp:path": "/exchange/cancelByCloid", "mcp:hidden": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cmd.Flags().Changed("cancels") && !flags.dryRun {
				return fmt.Errorf("required flag --cancels not set")
			}
			var cancels []sign.CancelCloidWire
			if bodyCancelsJSON != "" {
				if err := json.Unmarshal([]byte(bodyCancelsJSON), &cancels); err != nil {
					return fmt.Errorf("parsing --cancels JSON: %w", err)
				}
			}
			if len(cancels) == 0 && !flags.dryRun {
				return fmt.Errorf("--cancels must contain at least one {asset, cloid} entry")
			}
			action := sign.NewCancelCloidAction(cancels)

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
	cmd.Flags().StringVar(&bodyCancelsJSON, "cancels", "", "JSON array of {asset, cloid} cancellations")
	cmd.Flags().StringVar(&walletFlag, "wallet", "", "Override active wallet")
	return cmd
}
