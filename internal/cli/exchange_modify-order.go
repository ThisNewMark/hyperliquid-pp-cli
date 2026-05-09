// Hand-rewritten: typed ModifyAction + signed L1 POST.

package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"hyperliquid-pp-cli/internal/client"
	"hyperliquid-pp-cli/internal/sign"
)

func newExchangeModifyOrderCmd(flags *rootFlags) *cobra.Command {
	var bodyOid int
	var bodyOrderJSON string
	var walletFlag string

	cmd := &cobra.Command{
		Use:   "modify-order",
		Short: "Modify a single order (signed L1 action)",
		Example: `  hyperliquid exchange modify-order --oid 1234567890 --order '{"a":0,"b":true,"p":"40100","s":"0.001","r":false,"t":{"limit":{"tif":"Gtc"}}}'`,
		Annotations: map[string]string{
			"pp:endpoint": "exchange.modify-order",
			"pp:method":   "POST",
			"pp:path": "/exchange/modify",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cmd.Flags().Changed("oid") && !flags.dryRun {
				return fmt.Errorf("required flag --oid not set")
			}
			if !cmd.Flags().Changed("order") && !flags.dryRun {
				return fmt.Errorf("required flag --order not set")
			}
			var order sign.OrderWire
			if bodyOrderJSON != "" {
				if err := json.Unmarshal([]byte(bodyOrderJSON), &order); err != nil {
					return fmt.Errorf("parsing --order JSON: %w", err)
				}
			}
			action := sign.NewModifyAction(bodyOid, order)

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
	cmd.Flags().IntVar(&bodyOid, "oid", 0, "Order ID to modify")
	cmd.Flags().StringVar(&bodyOrderJSON, "order", "", "Replacement order JSON: {a,b,p,s,r,t,[c]}")
	cmd.Flags().StringVar(&walletFlag, "wallet", "", "Override active wallet")
	return cmd
}
