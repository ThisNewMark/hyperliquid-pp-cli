// Hand-rewritten: typed BatchModifyAction + signed L1 POST.

package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"hyperliquid-pp-cli/internal/client"
	"hyperliquid-pp-cli/internal/sign"
)

func newExchangeBatchModifyOrdersCmd(flags *rootFlags) *cobra.Command {
	var modifiesJSON string
	var walletFlag string

	cmd := &cobra.Command{
		Use:     "batch-modify-orders",
		Short:   "Modify multiple orders atomically (signed L1 action)",
		Example: `  hyperliquid exchange batch-modify-orders --modifies '[{"oid":1234,"order":{...}}]'`,
		Annotations: map[string]string{
			"pp:endpoint": "exchange.batch-modify-orders",
			"pp:method":   "POST",
			"pp:path": "/exchange/batchModify", "mcp:hidden": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cmd.Flags().Changed("modifies") && !flags.dryRun {
				return fmt.Errorf("required flag --modifies not set")
			}
			var modifies []sign.ModifyWire
			if modifiesJSON != "" {
				if err := json.Unmarshal([]byte(modifiesJSON), &modifies); err != nil {
					return fmt.Errorf("parsing --modifies JSON: %w", err)
				}
			}
			action := sign.NewBatchModifyAction(modifies)
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
	cmd.Flags().StringVar(&modifiesJSON, "modifies", "", "JSON array of {oid, order} modify entries")
	cmd.Flags().StringVar(&walletFlag, "wallet", "", "Override active wallet")
	return cmd
}
