// Wallet resolution for signed commands.
//
// Three wallet sources, in order of precedence (most → least specific):
//   1. --wallet flag on the command (overrides config)
//   2. config.Wallet ("agent" or "main")
//
// "agent" means: load $HYPERLIQUID_AGENT_KEY (env hex) or
// config.AgentKeyPath (default ~/.hyperliquid/agent.key) from disk.
// "main" means: load $HYPERLIQUID_MAIN_KEY (env hex) or config.MainKeyPath
// (no default — must be set).
//
// Some action types (approveAgent, approveBuilderFee, withdraw3, usdSend,
// spotSend) ALWAYS use the main wallet regardless of the active selection,
// because Hyperliquid rejects them when signed by an agent. Callers force
// this by passing WalletMain to walletKey().

package cli

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"os"

	"hyperliquid-pp-cli/internal/agent"
	"hyperliquid-pp-cli/internal/config"
)

type walletKind int

const (
	walletAuto walletKind = iota
	walletAgent
	walletMain
)

// walletKey resolves the active signing key. `forced` overrides config; pass
// walletAuto to honor config.Wallet, walletMain for main-only actions, or
// walletAgent to require the agent key.
func walletKey(cfg *config.Config, forced walletKind) (*ecdsa.PrivateKey, string, error) {
	kind := forced
	if kind == walletAuto {
		switch cfg.Wallet {
		case "main":
			kind = walletMain
		case "", "agent":
			kind = walletAgent
		default:
			return nil, "", fmt.Errorf("config wallet=%q is not 'agent' or 'main'", cfg.Wallet)
		}
	}

	switch kind {
	case walletAgent:
		return loadAgentKey(cfg)
	case walletMain:
		return loadMainKey(cfg)
	default:
		return nil, "", errors.New("walletKey: unreachable")
	}
}

func loadAgentKey(cfg *config.Config) (*ecdsa.PrivateKey, string, error) {
	if v := os.Getenv("HYPERLIQUID_AGENT_KEY_HEX"); v != "" {
		a, err := agent.LoadFromHexEnv("HYPERLIQUID_AGENT_KEY_HEX")
		if err != nil {
			return nil, "", err
		}
		return a.PrivateKey(), a.Address, nil
	}
	path := cfg.AgentKeyPath
	if path == "" {
		var err error
		path, err = agent.DefaultPath()
		if err != nil {
			return nil, "", err
		}
	}
	a, err := agent.Load(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, "", fmt.Errorf("no agent key at %s; run `hyperliquid agent generate` (then `hyperliquid agent approve` from your main wallet)", path)
		}
		return nil, "", err
	}
	return a.PrivateKey(), a.Address, nil
}

func loadMainKey(cfg *config.Config) (*ecdsa.PrivateKey, string, error) {
	if v := os.Getenv("HYPERLIQUID_MAIN_KEY"); v != "" {
		a, err := agent.LoadFromHexEnv("HYPERLIQUID_MAIN_KEY")
		if err != nil {
			return nil, "", err
		}
		return a.PrivateKey(), a.Address, nil
	}
	if cfg.MainKeyPath == "" {
		return nil, "", errors.New("main wallet not configured; set HYPERLIQUID_MAIN_KEY env var (raw 0x... hex) or main_key_path in config")
	}
	a, err := agent.Load(cfg.MainKeyPath)
	if err != nil {
		return nil, "", fmt.Errorf("loading main key from %s: %w", cfg.MainKeyPath, err)
	}
	return a.PrivateKey(), a.Address, nil
}
