// Hand-rewritten: typed UpdateLeverageAction + signed L1 POST.

package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"hyperliquid-pp-cli/internal/client"
	"hyperliquid-pp-cli/internal/sign"
)

func newExchangeUpdateLeverageCmd(flags *rootFlags) *cobra.Command {
	var asset int
	var isCross bool
	var leverage int
	var walletFlag string

	cmd := &cobra.Command{
		Use:   "update-leverage",
		Short: "Update cross/isolated leverage for an asset (signed L1 action)",
		Annotations: map[string]string{
			"pp:method":   "POST",
			"pp:path": "/exchange/updateLeverage",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, name := range []string{"asset", "leverage"} {
				if !cmd.Flags().Changed(name) && !flags.dryRun {
					return fmt.Errorf("required flag --%s not set", name)
				}
			}
			action := sign.NewUpdateLeverageAction(asset, isCross, leverage)

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
	cmd.Flags().BoolVar(&isCross, "is-cross", true, "Cross-margin (default true) or isolated (false)")
	cmd.Flags().IntVar(&leverage, "leverage", 1, "New leverage value")
	cmd.Flags().StringVar(&walletFlag, "wallet", "", "Override active wallet")
	return cmd
}
