// Hand-rewritten from the press-generated stub: place-order needs to build
// a typed PlaceOrderAction, sign it, and POST the signed envelope to
// /exchange. The press's generic Post helper would have hit a non-existent
// /exchange/order path with an unsigned body — neither correct nor functional
// against real Hyperliquid.

package cli

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"hyperliquid-pp-cli/internal/builder"
	"hyperliquid-pp-cli/internal/client"
	"hyperliquid-pp-cli/internal/sign"
)

// Compact alias to keep printSignedDryRun's signature short.
type ecdsaPrivateKey = ecdsa.PrivateKey

func timeNowMillis() int64 { return time.Now().UnixMilli() }

func newExchangePlaceOrderCmd(flags *rootFlags) *cobra.Command {
	var bodyBuilder string
	var bodyBuilderFeeBps int
	var bodyNoBuilder bool
	var bodyExpiresAfter int64
	var bodyGrouping string
	var bodyOrdersJSON string
	var bodyVaultAddress string
	var walletFlag string

	cmd := &cobra.Command{
		Use:     "place-order",
		Short:   "Place one or more orders. Builder code attached by default; --no-builder to opt out.",
		Example: "  hyperliquid exchange place-order --orders '[{\"a\":0,\"b\":true,\"p\":\"40000\",\"s\":\"0.01\",\"r\":false,\"t\":{\"limit\":{\"tif\":\"Gtc\"}}}]'",
		Annotations: map[string]string{
			"pp:method":   "POST",
			"pp:path": "/exchange/order",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cmd.Flags().Changed("orders") && !flags.dryRun {
				return fmt.Errorf("required flag --orders not set")
			}

			// Parse --orders JSON into typed OrderWire structs. Defensive: the
			// JSON tags on OrderWire match the wire format users will write
			// manually (a/b/p/s/r/t/c).
			var orders []sign.OrderWire
			if bodyOrdersJSON != "" {
				if err := json.Unmarshal([]byte(bodyOrdersJSON), &orders); err != nil {
					return fmt.Errorf("parsing --orders JSON: %w", err)
				}
			}
			if len(orders) == 0 && !flags.dryRun {
				return fmt.Errorf("--orders must contain at least one order")
			}

			// Resolve builder field. Use the same helper but adapt for typed
			// action: AttachToOrderBody operates on map[string]any, so we
			// piggy-back on it to get validation, then read back the result
			// into a *BuilderInfo.
			scratch := map[string]any{}
			if _, err := builder.AttachToOrderBody(scratch, bodyBuilder, bodyBuilderFeeBps, bodyNoBuilder); err != nil {
				return err
			}
			var builderInfo *sign.BuilderInfo
			if b, ok := scratch["builder"].(map[string]any); ok {
				builderInfo = &sign.BuilderInfo{
					B: b["b"].(string),
					F: b["f"].(int),
				}
			}

			grouping := bodyGrouping
			if grouping == "" {
				grouping = "na"
			}
			action := sign.NewPlaceOrderAction(orders, grouping, builderInfo)

			// Resolve signer.
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

			opts := client.SignOpts{
				IsMainnet: cfg.IsMainnet,
			}
			if bodyExpiresAfter != 0 {
				exp := uint64(bodyExpiresAfter)
				opts.ExpiresAfter = &exp
			}
			if bodyVaultAddress != "" {
				addr, err := sign.ParseAddress(bodyVaultAddress)
				if err != nil {
					return fmt.Errorf("--vault-address: %w", err)
				}
				opts.VaultAddress = &addr
			}

			if flags.dryRun {
				return printSignedDryRun(cmd, action, signerAddr, opts, builderInfo, key)
			}

			data, status, err := c.PostSignedL1(action, key, opts)
			if err != nil {
				return classifyAPIError(err, flags)
			}

			return printSignedResult(cmd, flags, "/exchange", data, status, action, signerAddr, builderInfo)
		},
	}
	cmd.Flags().StringVar(&bodyBuilder, "builder", builder.DefaultBuilderAddress, "Builder address to credit on this order (use 0x0 or --no-builder to opt out)")
	cmd.Flags().IntVar(&bodyBuilderFeeBps, "builder-fee-bps", builder.DefaultBuilderFeeBps, "Builder fee in tenths of basis points (10 = 1bp = 0.01%; server cap 100 perps, 1000 spot)")
	cmd.Flags().BoolVar(&bodyNoBuilder, "no-builder", false, "Do not attach any builder code to this order")
	cmd.Flags().Int64Var(&bodyExpiresAfter, "expires-after", 0, "Action expires after this ms epoch (0 = never)")
	cmd.Flags().StringVar(&bodyGrouping, "grouping", "na", "Grouping: na, normalTpsl, positionTpsl")
	cmd.Flags().StringVar(&bodyOrdersJSON, "orders", "", "JSON array of OrderWire objects: [{a,b,p,s,r,t,[c]}, ...]")
	cmd.Flags().StringVar(&bodyVaultAddress, "vault-address", "", "Vault address to route the action through")
	cmd.Flags().StringVar(&walletFlag, "wallet", "", "Override active wallet for this command: agent (default) or main")
	return cmd
}

// printSignedDryRun prints the typed action and (when a key is available)
// the actual signed envelope that WOULD be POSTed, without sending. With no
// key, it prints just the action shape — useful for inspecting builder-field
// passthrough without any wallet configured.
func printSignedDryRun(cmd *cobra.Command, action any, signerAddr string, opts client.SignOpts, builderInfo *sign.BuilderInfo, key *ecdsaPrivateKey) error {
	out := map[string]any{
		"target":  "/exchange (POST)",
		"signer":  signerAddr,
		"action":  action,
		"opts":    map[string]any{"is_mainnet": opts.IsMainnet},
		"builder": builderInfo,
	}
	if opts.VaultAddress != nil {
		out["opts"].(map[string]any)["vault_address"] = strings.ToLower(opts.VaultAddress.Hex())
	}
	if opts.ExpiresAfter != nil {
		out["opts"].(map[string]any)["expires_after"] = *opts.ExpiresAfter
	}

	if key == nil {
		out["note"] = "dry run — no signature (no key configured); set HYPERLIQUID_AGENT_KEY_HEX or run `hyperliquid agent generate` to also sign on dry-run"
	} else {
		// Compute the actual signature so the user can inspect what would go
		// onto the wire. This is the "AFTER build" check that the action +
		// builder field survive msgpack -> keccak -> EIP-712 -> ECDSA.
		nonce := opts.Nonce
		if nonce == 0 {
			nonce = uint64(timeNowMillis())
		}
		sig, err := sign.SignL1Action(key, action, opts.VaultAddress, nonce, opts.ExpiresAfter, opts.IsMainnet)
		if err != nil {
			return fmt.Errorf("signing dry-run envelope: %w", err)
		}
		// Recompute the action hash too — handy for cross-checking against
		// other implementations (e.g. the official Python SDK).
		actionHash, _ := sign.ActionHash(action, opts.VaultAddress, nonce, opts.ExpiresAfter)
		out["nonce"] = nonce
		out["action_hash"] = sign.HexBytes(actionHash)
		out["signature"] = sig
		out["note"] = "dry run — signature computed but request NOT sent"
	}

	enc, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), string(enc))
	return nil
}

// printSignedResult is the post-submit output. Mirrors the standard
// Cobra-tree envelope so --json/--agent users see consistent shape.
func printSignedResult(cmd *cobra.Command, flags *rootFlags, path string, data json.RawMessage, status int, action any, signerAddr string, builderInfo *sign.BuilderInfo) error {
	if flags.asJSON || !isTerminal(cmd.OutOrStdout()) {
		if flags.quiet {
			return nil
		}
		envelope := map[string]any{
			"action":   "post",
			"resource": "exchange",
			"path":     path,
			"status":   status,
			"success":  status >= 200 && status < 300,
			"signer":   signerAddr,
			"builder":  builderInfo,
		}
		if len(data) > 0 {
			var parsed any
			if err := json.Unmarshal(data, &parsed); err == nil {
				envelope["data"] = parsed
			}
		}
		enc, err := json.Marshal(envelope)
		if err != nil {
			return err
		}
		return printOutput(cmd.OutOrStdout(), json.RawMessage(enc), true)
	}
	return printOutputWithFlags(cmd.OutOrStdout(), data, flags)
}
