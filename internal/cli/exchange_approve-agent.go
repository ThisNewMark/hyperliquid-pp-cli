// Hand-rewritten: typed approveAgent action + signed user-signed POST.
// Note: also implemented as `agent approve` for ergonomics. This command is
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

func newExchangeApproveAgentCmd(flags *rootFlags) *cobra.Command {
	var agentAddress string
	var agentName string

	cmd := &cobra.Command{
		Use:   "approve-agent",
		Short: "Approve an agent (API) wallet for trading. Signed by your MAIN wallet.",
		Annotations: map[string]string{
			"pp:endpoint": "exchange.approve-agent",
			"pp:method":   "POST",
			"pp:path": "/exchange/approveAgent", "mcp:hidden": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cmd.Flags().Changed("agent-address") && !flags.dryRun {
				return fmt.Errorf("required flag --agent-address not set")
			}
			addr := strings.ToLower(strings.TrimSpace(agentAddress))
			action := map[string]any{
				"type":         "approveAgent",
				"agentAddress": addr,
				"agentName":    agentName,
			}
			primary, payloadTypes := sign.ApproveAgentSpec()

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
				"agent_address": addr,
				"agent_name":    agentName,
				"signer":        mainAddr,
				"status":        status,
				"success":       status >= 200 && status < 300,
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
	cmd.Flags().StringVar(&agentAddress, "agent-address", "", "Agent wallet 0x address to approve (required)")
	cmd.Flags().StringVar(&agentName, "agent-name", "hyperliquid-cli", "Agent display name (visible on-chain)")
	return cmd
}
