// `hyperliquid agent setup` — orchestrates the one-time on-chain
// authorization flow via the browser. The user signs both `approveAgent`
// and `approveBuilderFee` via MetaMask in a single page session; the CLI
// polls Hyperliquid until the builder approval lands and reports success.

package cli

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"hyperliquid-pp-cli/internal/agent"
	"hyperliquid-pp-cli/internal/builder"
)

// DefaultSetupURL is the static page that handles the browser-based signing
// flow. Override at runtime with --setup-url.
//
// Update this when the GitHub Pages site is live. For a repo at
// github.com/<user>/hyperliquid-pp-cli with Pages enabled on /docs, the URL is
//   https://<user>.github.io/hyperliquid-pp-cli/setup.html
const DefaultSetupURL = "https://thisnewmark.github.io/hyperliquid-pp-cli/setup.html"

func newAgentSetupCmd(flags *rootFlags) *cobra.Command {
	var setupURL string
	var noOpen bool
	var pollSeconds int
	var maxFeeRate string
	var builderAddrFlag string
	var userAddrFlag string

	cmd := &cobra.Command{
		Use:   "setup",
		Short: "One-shot browser-based setup: generates an agent and walks you through both on-chain approvals via your wallet",
		Long: `Generates a trading agent (if missing), then opens a setup page in your
browser. On that page, you'll sign two messages with your wallet (MetaMask,
Rabby, etc.) — one to authorize the agent, one to approve the builder fee.

Your wallet's private key is never exposed to the CLI or to the browser
beyond the wallet extension's normal signing UI.

After you sign both, this command polls Hyperliquid until the approval is
visible on-chain, then reports success. The whole flow takes about 60 seconds.`,
		Example: `  hyperliquid agent setup
  hyperliquid agent setup --max-fee-rate 0.05% --no-open`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			cfg := c.Config

			// 1. Ensure agent exists.
			path, err := agent.DefaultPath()
			if err != nil {
				return err
			}
			a, created, err := agent.LoadOrGenerate(path)
			if err != nil {
				return fmt.Errorf("preparing agent key: %w", err)
			}
			if created {
				fmt.Fprintf(cmd.OutOrStdout(), "✓ Generated new agent at %s\n", path)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "✓ Using existing agent at %s\n", path)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "  Agent address: %s\n\n", a.Address)

			// 2. Resolve params for the setup URL.
			builderAddr := builderAddrFlag
			if builderAddr == "" {
				builderAddr = builder.DefaultBuilderAddress
			}
			builderAddr = strings.ToLower(strings.TrimSpace(builderAddr))
			if maxFeeRate == "" {
				maxFeeRate = builder.DefaultMaxFeeRate
			}
			network := "mainnet"
			if !cfg.IsMainnet {
				network = "testnet"
			}

			// 3. Construct the URL.
			fullURL, err := buildSetupURL(setupURL, a.Address, builderAddr, maxFeeRate, network)
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Setup link:")
			fmt.Fprintln(cmd.OutOrStdout(), "  "+fullURL)
			fmt.Fprintln(cmd.OutOrStdout())

			// 4. Open browser unless suppressed.
			if !noOpen {
				if err := openInBrowser(fullURL); err != nil {
					fmt.Fprintf(cmd.OutOrStdout(), "  (couldn't auto-open: %v — copy the link above)\n\n", err)
				} else {
					fmt.Fprintln(cmd.OutOrStdout(), "  → Browser opened. Complete the two signatures there.")
					fmt.Fprintln(cmd.OutOrStdout())
				}
			}

			// 5. Poll Hyperliquid until approval lands. We can only poll if
			// the user gave us their main wallet address (the CLI never sees
			// it otherwise — the wallet stays in MetaMask).
			userAddr := strings.ToLower(strings.TrimSpace(userAddrFlag))
			if userAddr == "" {
				fmt.Fprintln(cmd.OutOrStdout(),
					"After signing both messages in the browser, verify with:\n"+
						"  hyperliquid builder status --user 0xYOUR_MAIN_ADDRESS\n"+
						"\n(Or rerun this command with --user 0xYOUR_MAIN_ADDRESS to auto-detect approval.)")
				return nil
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Waiting for approval to land on-chain (Ctrl-C to stop polling)…")
			deadline := time.Now().Add(time.Duration(pollSeconds) * time.Second)
			pollInterval := 5 * time.Second
			builderApproved := false
			for time.Now().Before(deadline) {
				ok, err := approvalCheck(c, userAddr, builderAddr)
				if err == nil && ok {
					builderApproved = true
					break
				}
				time.Sleep(pollInterval)
				fmt.Fprint(cmd.OutOrStdout(), ".")
			}
			fmt.Fprintln(cmd.OutOrStdout())

			if builderApproved {
				fmt.Fprintln(cmd.OutOrStdout(), "\n✓ Builder approval detected on-chain. You're done — start trading with `hyperliquid exchange place-order`.")
			} else {
				fmt.Fprintln(cmd.OutOrStdout(),
					"\nDidn't see the approval land within "+fmt.Sprintf("%ds", pollSeconds)+
						". If you signed both messages, give it another minute and verify with:\n"+
						"  hyperliquid builder status --user "+userAddr)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&setupURL, "setup-url", DefaultSetupURL, "Override the setup page URL (or set HYPERLIQUID_SETUP_URL)")
	cmd.Flags().BoolVar(&noOpen, "no-open", false, "Print the URL but don't auto-open the browser")
	cmd.Flags().IntVar(&pollSeconds, "poll-seconds", 120, "Seconds to keep polling Hyperliquid before exiting (only used if --user is set)")
	cmd.Flags().StringVar(&maxFeeRate, "max-fee-rate", "", "Max fee rate to approve (default: builder.DefaultMaxFeeRate)")
	cmd.Flags().StringVar(&builderAddrFlag, "builder", "", "Builder address (default: builder.DefaultBuilderAddress)")
	cmd.Flags().StringVar(&userAddrFlag, "user", "", "Your main wallet 0x address — when set, the CLI polls Hyperliquid until your signature lands and reports success")
	return cmd
}

func buildSetupURL(base, agent, builderAddr, maxFee, network string) (string, error) {
	u, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("invalid setup URL %q: %w", base, err)
	}
	q := u.Query()
	q.Set("agent", agent)
	q.Set("builder", builderAddr)
	q.Set("maxFee", maxFee)
	q.Set("network", network)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func openInBrowser(target string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", target)
	case "linux":
		cmd = exec.Command("xdg-open", target)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", target)
	default:
		return fmt.Errorf("unsupported OS %q for auto-open", runtime.GOOS)
	}
	return cmd.Start()
}

// approvalCheck queries /info {type:approvedBuilders, user:<addr>} and reports
// whether the builder appears in the list. Used by the polling loop after the
// browser flow completes.
func approvalCheck(c httpPoster, user, builderAddr string) (bool, error) {
	res, _, err := c.Post("/info", map[string]any{"type": "approvedBuilders", "user": user})
	if err != nil {
		return false, err
	}
	var list []string
	if err := json.Unmarshal(res, &list); err != nil {
		return false, err
	}
	for _, b := range list {
		if strings.EqualFold(b, builderAddr) {
			return true, nil
		}
	}
	return false, nil
}

type httpPoster interface {
	Post(path string, body any) (json.RawMessage, int, error)
}
