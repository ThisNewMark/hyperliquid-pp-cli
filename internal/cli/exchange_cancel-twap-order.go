// Hand-rewritten: typed TwapCancelAction + signed L1 POST.

package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"hyperliquid-pp-cli/internal/client"
	"hyperliquid-pp-cli/internal/sign"
)

func newExchangeCancelTwapOrderCmd(flags *rootFlags) *cobra.Command {
	var asset int
	var twapID int
	var walletFlag string

	cmd := &cobra.Command{
		Use:   "cancel-twap-order",
		Short: "Cancel a TWAP order by twapId (signed L1 action)",
		Annotations: map[string]string{
			"pp:endpoint": "exchange.cancel-twap-order",
			"pp:method":   "POST",
			"pp:path": "/exchange/twapCancel", "mcp:hidden": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, name := range []string{"asset", "twap-id"} {
				if !cmd.Flags().Changed(name) && !flags.dryRun {
					return fmt.Errorf("required flag --%s not set", name)
				}
			}
			action := sign.NewTwapCancelAction(asset, twapID)
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
	cmd.Flags().IntVar(&twapID, "twap-id", 0, "TWAP order ID to cancel")
	cmd.Flags().StringVar(&walletFlag, "wallet", "", "Override active wallet")
	return cmd
}
