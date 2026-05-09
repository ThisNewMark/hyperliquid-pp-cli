// Hand-rewritten: typed approveBuilderFee action + signed user-signed POST.
// Note: also implemented as `builder approve` for ergonomics. This command is
// the raw-flags form for users who want full control.

package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"hyperliquid-pp-cli/internal/client"
	"hyperliquid-pp-cli/internal/sign"
)

func newExchangeApproveBuilderFeeCmd(flags *rootFlags) *cobra.Command {
	var builderAddr string
	var maxFeeRate string

	cmd := &cobra.Command{
		Use:   "approve-builder-fee",
		Short: "Approve a builder address with max fee rate. Signed by your MAIN wallet.",
		Annotations: map[string]string{
			"pp:endpoint": "exchange.approve-builder-fee",
			"pp:method":   "POST",
			"pp:path": "/exchange/approveBuilderFee", "mcp:hidden": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, name := range []string{"builder", "max-fee-rate"} {
				if !cmd.Flags().Changed(name) && !flags.dryRun {
					return fmt.Errorf("required flag --%s not set", name)
				}
			}
			addr := strings.ToLower(strings.TrimSpace(builderAddr))
			action := map[string]any{
				"type":       "approveBuilderFee",
				"maxFeeRate": maxFeeRate,
				"builder":    addr,
			}
			primary, payloadTypes := sign.ApproveBuilderFeeSpec()

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
			result := map[string]any{
				"builder":      addr,
				"max_fee_rate": maxFeeRate,
				"signer":       mainAddr,
				"status":       status,
				"success":      status >= 200 && status < 300,
			}
			var parsed any
			if err := json.Unmarshal(data, &parsed); err == nil {
				result["data"] = parsed
			}
			enc, _ := json.MarshalIndent(result, "", "  ")
			fmt.Fprintln(cmd.OutOrStdout(), string(enc))
			return nil
		},
	}
	cmd.Flags().StringVar(&builderAddr, "builder", "", "Builder 0x address to approve (required)")
	cmd.Flags().StringVar(&maxFeeRate, "max-fee-rate", "", "Max fee rate as percentage string, e.g. 0.01% (required)")
	return cmd
}
