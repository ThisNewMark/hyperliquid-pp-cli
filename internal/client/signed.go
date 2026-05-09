// Hyperliquid-specific signed POST helper. Lives alongside the generated
// client.go but is not generated — touch this file freely.
//
// Hyperliquid has exactly two write paths: POST /info (read-only, no signing)
// and POST /exchange (signed). Per-action paths like /exchange/order that the
// press-generated commands target don't actually exist server-side; this
// helper routes everything to /exchange with the proper envelope.

package client

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"hyperliquid-pp-cli/internal/sign"
)

// SignOpts modifies the signed envelope. Zero value is fine for the common
// case (no vault, no expires-after, default nonce = current ms).
type SignOpts struct {
	// Nonce in milliseconds since epoch. If zero, time.Now().UnixMilli() is used.
	Nonce uint64

	// VaultAddress, if set, is included in both the action hash and the
	// envelope. Hyperliquid uses it to route the action to a vault's
	// account instead of the signer's own.
	VaultAddress *common.Address

	// ExpiresAfter, if set, makes the action invalid past that ms timestamp.
	ExpiresAfter *uint64

	// IsMainnet selects the phantom-agent source character. Set from
	// Config.IsMainnet at the call site.
	IsMainnet bool
}

// PostSignedL1 signs an L1 action with the given key and POSTs the full
// envelope to /exchange. Returns the raw response bytes and HTTP status.
func (c *Client) PostSignedL1(action any, key *ecdsa.PrivateKey, opts SignOpts) (json.RawMessage, int, error) {
	if opts.Nonce == 0 {
		opts.Nonce = uint64(time.Now().UnixMilli())
	}
	sig, err := sign.SignL1Action(key, action, opts.VaultAddress, opts.Nonce, opts.ExpiresAfter, opts.IsMainnet)
	if err != nil {
		return nil, 0, fmt.Errorf("sign L1 action: %w", err)
	}
	return c.postExchange(action, opts, sig)
}

// PostSignedUser signs a user-signed action (transfers, approveAgent,
// approveBuilderFee, withdraw, etc.) and POSTs to /exchange. The action map
// must already contain `signatureChainId` and `hyperliquidChain`.
func (c *Client) PostSignedUser(
	action map[string]any,
	primaryType string,
	payloadTypes []sign.ApiType,
	key *ecdsa.PrivateKey,
	opts SignOpts,
) (json.RawMessage, int, error) {
	if opts.Nonce == 0 {
		opts.Nonce = uint64(time.Now().UnixMilli())
	}
	// User-signed actions carry the nonce in the action body itself, not the
	// envelope. Stamp it if the caller didn't already.
	if _, hasNonce := action["nonce"]; !hasNonce {
		// Some user-signed actions use "time" instead of "nonce" — leave
		// "time" alone if the action already has it.
		if _, hasTime := action["time"]; !hasTime {
			action["nonce"] = opts.Nonce
		}
	}
	if _, ok := action["signatureChainId"]; !ok {
		action["signatureChainId"] = sign.SignatureChainIdMain
	}
	if _, ok := action["hyperliquidChain"]; !ok {
		if opts.IsMainnet {
			action["hyperliquidChain"] = "Mainnet"
		} else {
			action["hyperliquidChain"] = "Testnet"
		}
	}

	sig, err := sign.SignUserAction(key, action, primaryType, payloadTypes)
	if err != nil {
		return nil, 0, fmt.Errorf("sign user action: %w", err)
	}
	return c.postExchange(action, opts, sig)
}

func (c *Client) postExchange(action any, opts SignOpts, sig sign.Signature) (json.RawMessage, int, error) {
	envelope := map[string]any{
		"action":    action,
		"nonce":     opts.Nonce,
		"signature": sig,
	}
	if opts.VaultAddress != nil {
		envelope["vaultAddress"] = strings.ToLower(opts.VaultAddress.Hex())
	}
	if opts.ExpiresAfter != nil {
		envelope["expiresAfter"] = *opts.ExpiresAfter
	}
	return c.Post("/exchange", envelope)
}
