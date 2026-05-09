// Hand-rewritten: typed TwapOrderAction + signed L1 POST.

package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"hyperliquid-pp-cli/internal/client"
	"hyperliquid-pp-cli/internal/sign"
)

func newExchangePlaceTwapOrderCmd(flags *rootFlags) *cobra.Command {
	var asset int
	var isBuy bool
	var size string
	var reduceOnly bool
	var minutes int
	var randomize bool
	var walletFlag string

	cmd := &cobra.Command{
		Use:   "place-twap-order",
		Short: "Place a TWAP order over N minutes (signed L1 action)",
		Example: `  hyperliquid exchange place-twap-order --asset 0 --is-buy --size 0.01 --minutes 60`,
		Long: `Time-Weighted-Average-Price order. Hyperliquid splits the size into
slices and places them over --minutes. Optionally randomizes timing.`,
		Annotations: map[string]string{
			"pp:endpoint": "exchange.place-twap-order",
			"pp:method":   "POST",
			"pp:path": "/exchange/twapOrder", "mcp:hidden": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, name := range []string{"asset", "size", "minutes"} {
				if !cmd.Flags().Changed(name) && !flags.dryRun {
					return fmt.Errorf("required flag --%s not set", name)
				}
			}
			action := sign.NewTwapOrderAction(sign.TwapWire{
				Asset:      asset,
				IsBuy:      isBuy,
				Size:       size,
				ReduceOnly: reduceOnly,
				Minutes:    minutes,
				Randomize:  randomize,
			})
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
	cmd.Flags().BoolVar(&isBuy, "is-buy", false, "Direction: true=buy, false=sell")
	cmd.Flags().StringVar(&size, "size", "", "Total size to TWAP")
	cmd.Flags().BoolVar(&reduceOnly, "reduce-only", false, "Reduce-only flag")
	cmd.Flags().IntVar(&minutes, "minutes", 30, "Duration in minutes")
	cmd.Flags().BoolVar(&randomize, "randomize", false, "Randomize slice timing")
	cmd.Flags().StringVar(&walletFlag, "wallet", "", "Override active wallet")
	return cmd
}
