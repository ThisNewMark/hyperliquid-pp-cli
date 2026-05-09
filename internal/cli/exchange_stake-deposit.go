// Hand-rewritten: typed cDeposit + signed user-signed POST.

package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"hyperliquid-pp-cli/internal/client"
	"hyperliquid-pp-cli/internal/sign"
)

func newExchangeStakeDepositCmd(flags *rootFlags) *cobra.Command {
	var wei int64

	cmd := &cobra.Command{
		Use:   "stake-deposit",
		Short: "Stake HYPE into the staking layer (signed user action — MAIN wallet)",
		Annotations: map[string]string{
			"pp:endpoint": "exchange.stake-deposit",
			"pp:method":   "POST",
			"pp:path": "/exchange/stakeDeposit", "mcp:hidden": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cmd.Flags().Changed("wei") && !flags.dryRun {
				return fmt.Errorf("required flag --wei not set")
			}
			action := map[string]any{
				"type": "cDeposit",
				"wei":  uint64(wei),
			}
			primary, payloadTypes := sign.CDepositSpec()

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
			result := map[string]any{"wei": wei, "signer": mainAddr, "status": status, "success": status >= 200 && status < 300}
			var parsed any
			if err := json.Unmarshal(data, &parsed); err == nil {
				result["data"] = parsed
			}
			enc, _ := json.MarshalIndent(result, "", "  ")
			fmt.Fprintln(cmd.OutOrStdout(), string(enc))
			return nil
		},
	}
	cmd.Flags().Int64Var(&wei, "wei", 0, "Amount of HYPE in wei (required)")
	return cmd
}
