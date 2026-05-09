// Hand-rewritten: typed VaultTransferAction + signed L1 POST.

package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"hyperliquid-pp-cli/internal/client"
	"hyperliquid-pp-cli/internal/sign"
)

func newExchangeVaultTransferCmd(flags *rootFlags) *cobra.Command {
	var vault string
	var isDeposit bool
	var usd float64
	var walletFlag string

	cmd := &cobra.Command{
		Use:   "vault-transfer",
		Short: "Deposit to or withdraw from a vault (signed L1 action)",
		Annotations: map[string]string{
			"pp:endpoint": "exchange.vault-transfer",
			"pp:method":   "POST",
			"pp:path": "/exchange/vaultTransfer", "mcp:hidden": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, name := range []string{"vault-address", "usd"} {
				if !cmd.Flags().Changed(name) && !flags.dryRun {
					return fmt.Errorf("required flag --%s not set", name)
				}
			}
			action := sign.NewVaultTransferAction(vault, isDeposit, usd)
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
	cmd.Flags().StringVar(&vault, "vault-address", "", "Vault 0x address")
	cmd.Flags().BoolVar(&isDeposit, "is-deposit", true, "true=deposit to vault, false=withdraw")
	cmd.Flags().Float64Var(&usd, "usd", 0, "USDC amount")
	cmd.Flags().StringVar(&walletFlag, "wallet", "", "Override active wallet")
	return cmd
}
