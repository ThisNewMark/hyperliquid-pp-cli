// Hand-rewritten: typed tokenDelegate + signed user-signed POST.

package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"hyperliquid-pp-cli/internal/client"
	"hyperliquid-pp-cli/internal/sign"
)

func newExchangeDelegateTokenCmd(flags *rootFlags) *cobra.Command {
	var validator string
	var isUndelegate bool
	var wei int64

	cmd := &cobra.Command{
		Use:   "delegate-token",
		Short: "Delegate (or undelegate) staked HYPE to a validator (signed user action — MAIN wallet)",
		Long:  "1-day lockup per delegation. Pass --is-undelegate to reverse.",
		Annotations: map[string]string{
			"pp:endpoint": "exchange.delegate-token",
			"pp:method":   "POST",
			"pp:path": "/exchange/delegateToken", "mcp:hidden": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, name := range []string{"validator", "wei"} {
				if !cmd.Flags().Changed(name) && !flags.dryRun {
					return fmt.Errorf("required flag --%s not set", name)
				}
			}
			validatorAddr := strings.ToLower(strings.TrimSpace(validator))
			action := map[string]any{
				"type":         "tokenDelegate",
				"validator":    validatorAddr,
				"wei":          uint64(wei),
				"isUndelegate": isUndelegate,
			}
			primary, payloadTypes := sign.TokenDelegateSpec()

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
			result := map[string]any{"validator": validatorAddr, "is_undelegate": isUndelegate, "wei": wei, "signer": mainAddr, "status": status, "success": status >= 200 && status < 300}
			var parsed any
			if err := json.Unmarshal(data, &parsed); err == nil {
				result["data"] = parsed
			}
			enc, _ := json.MarshalIndent(result, "", "  ")
			fmt.Fprintln(cmd.OutOrStdout(), string(enc))
			return nil
		},
	}
	cmd.Flags().StringVar(&validator, "validator", "", "Validator 0x address (required)")
	cmd.Flags().BoolVar(&isUndelegate, "is-undelegate", false, "true to undelegate, false to delegate")
	cmd.Flags().Int64Var(&wei, "wei", 0, "Amount in wei (required)")
	return cmd
}
