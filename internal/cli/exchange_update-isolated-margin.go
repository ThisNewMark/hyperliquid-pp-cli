// Hand-rewritten: typed UpdateIsolatedMarginAction + signed L1 POST.

package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"hyperliquid-pp-cli/internal/client"
	"hyperliquid-pp-cli/internal/sign"
)

func newExchangeUpdateIsolatedMarginCmd(flags *rootFlags) *cobra.Command {
	var asset int
	var isBuy bool
	var ntli int
	var walletFlag string

	cmd := &cobra.Command{
		Use:   "update-isolated-margin",
		Short: "Add/remove isolated margin on an asset (signed L1 action)",
		Long: `--ntli is the change in 6-decimal USDC notional (e.g., 1000000 = $1).
Positive adds margin, negative removes.`,
		Annotations: map[string]string{
			"pp:endpoint": "exchange.update-isolated-margin",
			"pp:method":   "POST",
			"pp:path": "/exchange/updateIsolatedMargin", "mcp:hidden": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, name := range []string{"asset", "ntli"} {
				if !cmd.Flags().Changed(name) && !flags.dryRun {
					return fmt.Errorf("required flag --%s not set", name)
				}
			}
			action := sign.NewUpdateIsolatedMarginAction(asset, isBuy, ntli)
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
	cmd.Flags().IntVar(&asset, "asset", 0, "Asset index")
	cmd.Flags().BoolVar(&isBuy, "is-buy", true, "Side: true for long position, false for short")
	cmd.Flags().IntVar(&ntli, "ntli", 0, "Change in 6-decimal USDC notional (1000000 = $1)")
	cmd.Flags().StringVar(&walletFlag, "wallet", "", "Override active wallet")
	return cmd
}
