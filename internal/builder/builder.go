// Package builder owns the CLI's relationship to Hyperliquid builder codes.
//
// Hyperliquid lets any address ("builder") attach a fee field to orders that a
// user has pre-approved via the on-chain ApproveBuilderFee action. The fee is
// expressed as `f` in tenths of basis points; e.g. f=10 means 1bp = 0.01%.
// Server caps: 0.1% on perps and 1% on spot.
//
// This package centralizes:
//   - the default builder address shipped with this CLI
//   - the default fee in tenths of bps
//   - the helper that attaches the builder field to an outgoing order body
//   - the helper that constructs the ApproveBuilderFee body
package builder

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// DefaultBuilderAddress is the address that receives builder fees by default
// for orders placed through this CLI. Lowercased; Hyperliquid expects the
// builder address lowercased on the wire.
const DefaultBuilderAddress = "0xc8f0cd137e28f717a20f810b46926f92978bbcfa"

// DefaultBuilderFeeBps is the fee, expressed in **tenths of basis points**,
// applied to orders placed through this CLI when the builder field is on.
// 10 = 1 basis point = 0.01%. Server cap is 100 (= 0.1%) on perps and 1000
// (= 1%) on spot.
const DefaultBuilderFeeBps = 10

// DefaultMaxFeeRate is the human-readable percentage string used when the
// builder approve flow runs without an explicit --max-fee-rate flag. This
// must be ≥ DefaultBuilderFeeBps (after unit conversion) or the server will
// reject orders carrying our fee. Set to match DefaultBuilderFeeBps exactly
// (10 tenths-of-bps = 0.01%), so a fresh `builder approve` permits the
// default fee with no headroom. To allow future fee bumps without
// re-approval, pass a higher --max-fee-rate (server cap is 0.1% perps).
const DefaultMaxFeeRate = "0.01%"

// SentinelDisable is the value that, when passed to --builder, signals an
// explicit opt-out without using --no-builder. Useful for transparency in
// scripted invocations.
const SentinelDisable = "0x0"

var addressRe = regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`)

// NormalizeAddress lowercases the input and validates the 0x... shape.
// Returns "" if the input is the SentinelDisable value.
func NormalizeAddress(addr string) (string, error) {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return "", errors.New("address is empty")
	}
	if strings.EqualFold(addr, SentinelDisable) {
		return "", nil
	}
	addr = strings.ToLower(addr)
	if !addressRe.MatchString(addr) {
		return "", fmt.Errorf("invalid 0x address %q (must be 0x + 40 hex chars)", addr)
	}
	return addr, nil
}

// FeeBpsField returns the JSON shape Hyperliquid expects for the builder
// field on an order: { "b": <addr>, "f": <fee in tenths of bps> }.
func FeeBpsField(addr string, feeBps int) map[string]any {
	return map[string]any{
		"b": addr,
		"f": feeBps,
	}
}

// AttachToOrderBody mutates a /exchange/order request body to carry the
// builder field, applying the CLI's defaults and respecting an explicit
// opt-out. Returns whether the field was actually set so callers can report
// the choice in --dry-run / --json output.
//
// Precedence:
//   - If noBuilder is true, do nothing and return false.
//   - If addr is the SentinelDisable ("0x0"), do nothing.
//   - Otherwise validate addr is a 0x...40-hex address and attach.
func AttachToOrderBody(body map[string]any, addr string, feeBps int, noBuilder bool) (attached bool, err error) {
	if body == nil {
		return false, errors.New("nil body")
	}
	if noBuilder {
		// User asked for full transparency — drop any builder field that may
		// have been pre-set by stdin, profile, or earlier wiring.
		delete(body, "builder")
		return false, nil
	}
	normalized, err := NormalizeAddress(addr)
	if err != nil {
		return false, err
	}
	if normalized == "" {
		// SentinelDisable.
		delete(body, "builder")
		return false, nil
	}
	if feeBps < 0 {
		return false, fmt.Errorf("--builder-fee-bps must be ≥ 0, got %d", feeBps)
	}
	body["builder"] = FeeBpsField(normalized, feeBps)
	return true, nil
}

// ApproveBuilderFeeBody returns the action body that the on-chain
// approveBuilderFee call sends to /exchange. Caller must EIP-712 sign with
// the user's MAIN wallet (not an agent) before submission.
//
// TOMORROW: when the signer lands, ensure this body is wrapped in the
// HyperliquidSignTransaction:ApproveBuilderFee EIP-712 typed-data envelope.
func ApproveBuilderFeeBody(builder string, maxFeeRate string, hyperliquidChain string, signatureChainId string) (map[string]any, error) {
	addr, err := NormalizeAddress(builder)
	if err != nil {
		return nil, err
	}
	if addr == "" {
		return nil, errors.New("cannot approve the disabled sentinel address; pass a real builder")
	}
	if maxFeeRate == "" {
		maxFeeRate = DefaultMaxFeeRate
	}
	if hyperliquidChain == "" {
		hyperliquidChain = "Mainnet"
	}
	if signatureChainId == "" {
		// Arbitrum One. Hyperliquid uses Arbitrum chain IDs in the signature
		// envelope even though L1 actions use chainId 1337.
		signatureChainId = "0xa4b1"
	}
	return map[string]any{
		"type":             "approveBuilderFee",
		"hyperliquidChain": hyperliquidChain,
		"signatureChainId": signatureChainId,
		"maxFeeRate":       maxFeeRate,
		"builder":          addr,
	}, nil
}
