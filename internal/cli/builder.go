// Top-level `builder` command group.
//
// Hand-authored. Wraps Hyperliquid's approveBuilderFee action and the
// /info {approvedBuilders, maxBuilderFee} read endpoints.

package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"hyperliquid-pp-cli/internal/builder"
	"hyperliquid-pp-cli/internal/client"
	"hyperliquid-pp-cli/internal/sign"
)

func newBuilderCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "builder",
		Short: "Manage builder-code approval for this CLI's default builder",
		Long: `Manage the on-chain builder-code relationship between your wallet and this CLI's default builder address.

Builder codes let any client attach a fee field to orders that the user has
pre-approved on-chain. This CLI ships with a default builder address; you can
opt-out per-order with --no-builder, or fully opt-out by setting --builder 0x0.

See "Builder Code Transparency" in the README for details on how to verify
the address and how the fee is calculated.`,
	}
	cmd.AddCommand(newBuilderApproveCmd(flags))
	cmd.AddCommand(newBuilderStatusCmd(flags))
	cmd.AddCommand(newBuilderRevokeCmd(flags))
	return cmd
}

func newBuilderApproveCmd(flags *rootFlags) *cobra.Command {
	var addr string
	var maxFeeRate string

	cmd := &cobra.Command{
		Use:   "approve",
		Short: "Approve a builder address on-chain (signed by your MAIN wallet)",
		Long: `Sends an approveBuilderFee action authorizing --builder to charge up to
--max-fee-rate per order signed by your wallet.

Server caps: 0.1% perps, 1% spot. MUST be signed by the main depositing
wallet — agents cannot self-authorize builder fees.`,
		Example: `  hyperliquid builder approve
  hyperliquid builder approve --builder 0xabc... --max-fee-rate 0.005%`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			cfg := c.Config

			normalized, err := builder.NormalizeAddress(addr)
			if err != nil {
				return err
			}
			if normalized == "" {
				return fmt.Errorf("--builder cannot be the disabled sentinel for approve")
			}

			action := map[string]any{
				"type":       "approveBuilderFee",
				"maxFeeRate": maxFeeRate,
				"builder":    normalized,
			}
			primary, payloadTypes := sign.ApproveBuilderFeeSpec()

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
				"builder":      normalized,
				"max_fee_rate": maxFeeRate,
				"signer":       mainAddr,
				"status":       status,
				"success":      status >= 200 && status < 300,
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
	cmd.Flags().StringVar(&addr, "builder", builder.DefaultBuilderAddress, "Builder address to approve (defaults to this CLI's address)")
	cmd.Flags().StringVar(&maxFeeRate, "max-fee-rate", builder.DefaultMaxFeeRate, "Max fee rate as a percentage string, e.g. 0.01%")
	return cmd
}

func newBuilderStatusCmd(flags *rootFlags) *cobra.Command {
	var user string
	var addr string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show this CLI's builder-fee approval state for a user",
		Long: `Reads /info {type:approvedBuilders} and /info {type:maxBuilderFee} to show
whether --user has approved --builder, and at what fee cap.

If --builder is omitted, the CLI's default builder address is used.`,
		Example: `  hyperliquid builder status --user 0xabc...`,
		RunE: func(cmd *cobra.Command, args []string) error {
			normalized, err := builder.NormalizeAddress(addr)
			if err != nil {
				return err
			}
			if normalized == "" {
				return fmt.Errorf("builder address resolved to the disabled sentinel; pass --builder explicitly")
			}
			if user == "" {
				return fmt.Errorf("--user is required")
			}

			c, err := flags.newClient()
			if err != nil {
				return err
			}

			approved, _, err := c.Post("/info", map[string]any{"type": "approvedBuilders", "user": user})
			if err != nil {
				return classifyAPIError(err, flags)
			}
			capRaw, _, err := c.Post("/info", map[string]any{"type": "maxBuilderFee", "user": user, "builder": normalized})
			if err != nil {
				return classifyAPIError(err, flags)
			}

			var approvedList []string
			_ = json.Unmarshal(approved, &approvedList)
			var capInt int
			_ = json.Unmarshal(capRaw, &capInt)

			isApproved := false
			for _, a := range approvedList {
				if a == normalized {
					isApproved = true
					break
				}
			}

			result := map[string]any{
				"user":               user,
				"builder":            normalized,
				"approved":           isApproved,
				"max_fee_tenths_bps": capInt,
				"max_fee_pct":        fmt.Sprintf("%.4f%%", float64(capInt)/1000.0),
				"approved_builders":  approvedList,
			}
			out, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(out))
			return nil
		},
	}
	cmd.Flags().StringVar(&user, "user", "", "User address to query (required)")
	cmd.Flags().StringVar(&addr, "builder", builder.DefaultBuilderAddress, "Builder address to check (defaults to this CLI's address)")
	return cmd
}

func newBuilderRevokeCmd(flags *rootFlags) *cobra.Command {
	var addr string

	cmd := &cobra.Command{
		Use:   "revoke",
		Short: "Revoke a builder approval (sets maxFeeRate to 0%)",
		Example: `  hyperliquid builder revoke --builder 0xc8f0cd137e28f717a20f810b46926f92978bbcfa`,
		Long: `Issues an approveBuilderFee action with maxFeeRate=0%, which has the
effect of revoking a prior approval. Like approve, must be signed by your
MAIN wallet.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			cfg := c.Config
			normalized, err := builder.NormalizeAddress(addr)
			if err != nil {
				return err
			}
			if normalized == "" {
				return fmt.Errorf("--builder cannot be the disabled sentinel for revoke")
			}

			action := map[string]any{
				"type":       "approveBuilderFee",
				"maxFeeRate": "0%",
				"builder":    normalized,
			}
			primary, payloadTypes := sign.ApproveBuilderFeeSpec()

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
				"builder": normalized,
				"effect":  "revokes prior approval (maxFeeRate=0%)",
				"signer":  mainAddr,
				"status":  status,
				"success": status >= 200 && status < 300,
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
	cmd.Flags().StringVar(&addr, "builder", builder.DefaultBuilderAddress, "Builder address to revoke (defaults to this CLI's address)")
	return cmd
}
