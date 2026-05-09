// Top-level `agent` command group: manages the CLI's agent wallet (Mode 2).
//
// Flow:
//   hyperliquid agent generate         — make a fresh keypair, persist 0600
//   hyperliquid agent approve          — approveAgent on-chain (signs with main wallet)
//   hyperliquid agent status           — show what's persisted + on-chain approval state
//   hyperliquid agent revoke           — wipe local agent (does NOT undo on-chain approval)
//
// approveAgent requires the user's MAIN wallet — the daily agent key cannot
// authorize itself.

package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"hyperliquid-pp-cli/internal/agent"
	"hyperliquid-pp-cli/internal/client"
	"hyperliquid-pp-cli/internal/sign"
)

func newAgentCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent",
		Short: "Manage this CLI's agent wallet (recommended trading key)",
		Long: `Hyperliquid lets a "main" wallet authorize one or more "agent" wallets to
trade on its behalf. Agents can place/cancel orders but CANNOT call withdraw,
transfer, or approve more agents — so a leaked agent key has bounded blast
radius.

This CLI defaults to signing daily trading actions with an agent wallet, kept
at ~/.hyperliquid/agent.key (mode 0600). One-time setup:

    hyperliquid agent generate         # makes a fresh agent
    hyperliquid agent approve          # main wallet authorizes the agent on-chain
    hyperliquid exchange place-order   # subsequent trades sign with the agent`,
	}
	cmd.AddCommand(newAgentGenerateCmd(flags))
	cmd.AddCommand(newAgentApproveCmd(flags))
	cmd.AddCommand(newAgentSetupCmd(flags))
	cmd.AddCommand(newAgentStatusCmd(flags))
	cmd.AddCommand(newAgentRevokeCmd(flags))
	return cmd
}

func newAgentGenerateCmd(flags *rootFlags) *cobra.Command {
	var force bool
	var customPath string

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate a fresh agent keypair and persist it to disk (mode 0600)",
		RunE: func(cmd *cobra.Command, args []string) error {
			path := customPath
			if path == "" {
				p, err := agent.DefaultPath()
				if err != nil {
					return err
				}
				path = p
			}
			a, err := agent.Generate()
			if err != nil {
				return err
			}
			if err := a.Save(path, force); err != nil {
				return err
			}

			out := map[string]any{
				"agent_address": a.Address,
				"path":          path,
				"next_step":     fmt.Sprintf("hyperliquid agent approve --agent %s", a.Address),
				"warning":       "this agent is NOT authorized yet. The next step signs an approveAgent action with your MAIN wallet.",
			}
			enc, err := json.MarshalIndent(out, "", "  ")
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(enc))
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite an existing agent key file (destroys the previous agent — its on-chain approval is left dangling)")
	cmd.Flags().StringVar(&customPath, "path", "", "Override key file path (default: $HOME/.hyperliquid/agent.key or $HYPERLIQUID_AGENT_KEY)")
	return cmd
}

func newAgentApproveCmd(flags *rootFlags) *cobra.Command {
	var agentAddrFlag string
	var agentName string

	cmd := &cobra.Command{
		Use:   "approve",
		Short: "Approve the local agent address on-chain (signs with your MAIN wallet)",
		Example: `  hyperliquid agent approve
  hyperliquid agent approve --agent 0xabc... --name my-cli`,
		Long: `Sends an approveAgent action to /exchange. This authorizes the named agent
address to place/cancel orders on behalf of the signer's account.

Hyperliquid limits: 1 unnamed + 3 named agents per account, plus 2 per
sub-account. If you've reached the cap, revoke an old agent first.

Requires the MAIN wallet key — set HYPERLIQUID_MAIN_KEY (raw 0x... hex) or
main_key_path in config. The CLI never persists the main key.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			cfg := c.Config

			// Resolve which agent address we're approving.
			addr := strings.ToLower(strings.TrimSpace(agentAddrFlag))
			if addr == "" {
				p, err := agent.DefaultPath()
				if err != nil {
					return err
				}
				a, err := agent.Load(p)
				if err != nil {
					return fmt.Errorf("no --agent supplied and no local agent at %s: %w", p, err)
				}
				addr = a.Address
			}
			if !strings.HasPrefix(addr, "0x") || len(addr) != 42 {
				return fmt.Errorf("agent address %q does not look like a 0x...40-hex address", addr)
			}

			action := map[string]any{
				"type":         "approveAgent",
				"agentAddress": addr,
				"agentName":    agentName,
			}
			primary, payloadTypes := sign.ApproveAgentSpec()

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
			enc, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(enc))
			return nil
		},
	}
	cmd.Flags().StringVar(&agentAddrFlag, "agent", "", "Agent address to approve (default: local agent at ~/.hyperliquid/agent.key)")
	cmd.Flags().StringVar(&agentName, "name", "hyperliquid-cli", "Agent name (visible on-chain)")
	return cmd
}

func newAgentStatusCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show local agent address and (optionally) on-chain approval state",
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := agent.DefaultPath()
			if err != nil {
				return err
			}
			out := map[string]any{
				"path": path,
			}
			if a, err := agent.Load(path); err == nil {
				out["agent_address"] = a.Address
				out["status"] = "loaded"
			} else {
				out["status"] = "missing"
				out["error"] = err.Error()
			}
			enc, err := json.MarshalIndent(out, "", "  ")
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(enc))
			return nil
		},
	}
	return cmd
}

func newAgentRevokeCmd(flags *rootFlags) *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "revoke",
		Short: "Wipe the local agent key file (does NOT undo on-chain approval — Hyperliquid agents are valid until manually rotated)",
		Example: `  hyperliquid agent revoke --force`,
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := agent.DefaultPath()
			if err != nil {
				return err
			}
			if !force {
				return fmt.Errorf("refusing to delete %s without --force; the on-chain approval will outlast this rotation, and the new local agent will need a fresh `agent approve` from the main wallet", path)
			}
			// Best-effort: ignore "already gone".
			_ = removeIfExists(path)
			fmt.Fprintln(cmd.OutOrStdout(), "removed:", path)
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "Confirm deletion of the local agent key file")
	return cmd
}

func removeIfExists(path string) error {
	if _, err := agent.Load(path); err != nil {
		return nil
	}
	return removeFile(path)
}

func printApproveDryRun(cmd *cobra.Command, action map[string]any, primary string, types []sign.ApiType) error {
	out := map[string]any{
		"target":        "/exchange (POST, user-signed)",
		"primary_type":  primary,
		"payload_types": types,
		"action":        action,
		"signing":       "MAIN wallet only — agents cannot self-authorize",
		"note":          "dry run — no signature computed, no request sent",
	}
	enc, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), string(enc))
	return nil
}
