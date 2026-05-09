// Hand-rewritten from the press-generated stub: cancel needs to build a
// typed CancelOidAction, sign it as L1, and POST the signed envelope.

package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"hyperliquid-pp-cli/internal/client"
	"hyperliquid-pp-cli/internal/sign"
)

func newExchangeCancelOrdersCmd(flags *rootFlags) *cobra.Command {
	var bodyCancelsJSON string
	var walletFlag string

	cmd := &cobra.Command{
		Use:     "cancel-orders",
		Short:   "Cancel orders by oid (signed L1 action)",
		Example: "  hyperliquid exchange cancel-orders --cancels '[{\"a\":0,\"o\":1234567890}]'",
		Annotations: map[string]string{
			"pp:method":   "POST",
			"pp:path": "/exchange/cancel",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cmd.Flags().Changed("cancels") && !flags.dryRun {
				return fmt.Errorf("required flag --cancels not set")
			}
			var cancels []sign.CancelOidWire
			if bodyCancelsJSON != "" {
				if err := json.Unmarshal([]byte(bodyCancelsJSON), &cancels); err != nil {
					return fmt.Errorf("parsing --cancels JSON: %w", err)
				}
			}
			if len(cancels) == 0 && !flags.dryRun {
				return fmt.Errorf("--cancels must contain at least one {a, o} entry")
			}
			action := sign.NewCancelOidAction(cancels)

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
	cmd.Flags().StringVar(&bodyCancelsJSON, "cancels", "", "JSON array of {a:assetIdx, o:oid} cancellations")
	cmd.Flags().StringVar(&walletFlag, "wallet", "", "Override active wallet for this command: agent (default) or main")
	return cmd
}
