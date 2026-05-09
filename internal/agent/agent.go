// Package agent manages this CLI's agent wallet — a derivative ECDSA
// keypair that the user authorizes once via approveAgent, then uses to sign
// daily trading actions instead of their main wallet.
//
// Threat model: if your laptop is compromised, an agent key gets stolen.
// Worst case: attacker can place/cancel orders. They CANNOT call withdraw3,
// usdSend, spotSend, approveAgent, or approveBuilderFee — those are
// server-side restricted to main wallets. So leaks are bounded.
//
// Storage: keys live in $HYPERLIQUID_HOME/agent.key (default
// $HOME/.hyperliquid/agent.key) at mode 0600. The file holds the 64-char hex
// of the private key, a single newline, then the 0x-prefixed address.
package agent

import (
	"bufio"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
)

// Agent owns a keypair and reports its public address.
type Agent struct {
	Address string // lowercase 0x...
	key     *ecdsa.PrivateKey
}

// PrivateKey returns the underlying ECDSA private key.
func (a *Agent) PrivateKey() *ecdsa.PrivateKey {
	return a.key
}

// DefaultPath is where the agent key lives when no path override is given.
func DefaultPath() (string, error) {
	if v := os.Getenv("HYPERLIQUID_AGENT_KEY"); v != "" {
		return v, nil
	}
	if v := os.Getenv("HYPERLIQUID_HOME"); v != "" {
		return filepath.Join(v, "agent.key"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".hyperliquid", "agent.key"), nil
}

// Generate creates a fresh ECDSA keypair (secp256k1) suitable for use as an
// agent wallet on Hyperliquid.
func Generate() (*Agent, error) {
	key, err := crypto.GenerateKey()
	if err != nil {
		return nil, fmt.Errorf("generate key: %w", err)
	}
	return &Agent{
		Address: strings.ToLower(crypto.PubkeyToAddress(key.PublicKey).Hex()),
		key:     key,
	}, nil
}

// Save persists the agent to disk at mode 0600. Refuses to overwrite an
// existing file unless overwrite is true.
func (a *Agent) Save(path string, overwrite bool) error {
	if path == "" {
		return errors.New("Save: empty path")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(path), err)
	}
	flag := os.O_WRONLY | os.O_CREATE | os.O_EXCL
	if overwrite {
		flag = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	}
	f, err := os.OpenFile(path, flag, 0o600)
	if err != nil {
		if os.IsExist(err) && !overwrite {
			return fmt.Errorf("agent key already exists at %s; use --force to overwrite (this destroys the existing agent and any on-chain approval is left dangling)", path)
		}
		return err
	}
	defer f.Close()

	priv := hex.EncodeToString(crypto.FromECDSA(a.key))
	if _, err := io.WriteString(f, priv+"\n"+a.Address+"\n"); err != nil {
		return err
	}
	return nil
}

// Load reads an agent key from disk. Refuses to load if the file mode is more
// permissive than 0600 — protects against accidental world-readable keys.
func Load(path string) (*Agent, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if mode := info.Mode().Perm(); mode&0o077 != 0 {
		return nil, fmt.Errorf("agent key file %s has insecure mode %o; expected 0600 (run: chmod 600 %s)", path, mode, path)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 256), 256)
	if !scanner.Scan() {
		return nil, fmt.Errorf("agent key file %s is empty", path)
	}
	hexKey := strings.TrimSpace(scanner.Text())
	hexKey = strings.TrimPrefix(hexKey, "0x")
	if len(hexKey) != 64 {
		return nil, fmt.Errorf("agent key file %s: expected 64-char hex private key, got %d chars", path, len(hexKey))
	}
	keyBytes, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, fmt.Errorf("agent key file %s: invalid hex: %w", path, err)
	}
	priv, err := crypto.ToECDSA(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("agent key file %s: %w", path, err)
	}
	addr := strings.ToLower(crypto.PubkeyToAddress(priv.PublicKey).Hex())
	return &Agent{Address: addr, key: priv}, nil
}

// LoadOrGenerate returns the agent at path, creating one if missing.
// Newly-generated agents are saved before returning. The boolean reports
// whether a new agent was just created (true) or an existing one was loaded
// (false).
func LoadOrGenerate(path string) (*Agent, bool, error) {
	if _, err := os.Stat(path); err == nil {
		a, err := Load(path)
		if err != nil {
			return nil, false, err
		}
		return a, false, nil
	} else if !os.IsNotExist(err) {
		return nil, false, err
	}
	a, err := Generate()
	if err != nil {
		return nil, false, err
	}
	if err := a.Save(path, false); err != nil {
		return nil, false, err
	}
	return a, true, nil
}

// LoadFromHexEnv reads a key directly from the named env var. Useful for
// loading a "main" wallet without ever writing it to disk.
func LoadFromHexEnv(envVar string) (*Agent, error) {
	v := strings.TrimSpace(os.Getenv(envVar))
	if v == "" {
		return nil, fmt.Errorf("env var %s is empty", envVar)
	}
	v = strings.TrimPrefix(v, "0x")
	if len(v) != 64 {
		return nil, fmt.Errorf("env var %s: expected 64-char hex private key, got %d chars", envVar, len(v))
	}
	keyBytes, err := hex.DecodeString(v)
	if err != nil {
		return nil, fmt.Errorf("env var %s: invalid hex: %w", envVar, err)
	}
	priv, err := crypto.ToECDSA(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("env var %s: %w", envVar, err)
	}
	return &Agent{
		Address: strings.ToLower(crypto.PubkeyToAddress(priv.PublicKey).Hex()),
		key:     priv,
	}, nil
}
