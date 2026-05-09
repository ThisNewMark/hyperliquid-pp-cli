// Package sign implements Hyperliquid's two action-signing schemes.
//
// L1 actions (orders, cancels, modifies, leverage updates, twap, schedule
// cancel) use a "phantom agent" construction:
//
//  1. Encode the action object with msgpack (field order matters; this
//     package uses Go struct definition order).
//  2. Append nonce (big-endian uint64).
//  3. Append 0x00 if no vault, else 0x01 || 20-byte address.
//  4. If expires-after is set, append 0x00 || big-endian uint64.
//     (Yes, 0x00 — this matches the official Python SDK's behavior. Do not
//     change without the spec changing first.)
//  5. keccak256 the whole thing → connectionId.
//  6. EIP-712 sign over Agent{source, connectionId} with domain
//     {Exchange, "1", chainId 1337, 0x0...0}.
//
// User-signed actions (transfers, withdrawals, approveAgent, approveBuilderFee,
// staking) skip the phantom-agent step and instead EIP-712-sign the action
// object directly with domain {HyperliquidSignTransaction, "1",
// chainId from action.signatureChainId, 0x0...0} and a primary type per the
// action.
//
// Reference: hyperliquid-dex/hyperliquid-python-sdk hyperliquid/utils/signing.py
package sign

import (
	"crypto/ecdsa"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	ethmath "github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/vmihailenco/msgpack/v5"
)

// Signature is the Hyperliquid wire format for a signed action.
type Signature struct {
	R string `json:"r"`
	S string `json:"s"`
	V byte   `json:"v"`
}

// ApiType is the EIP-712 typed-data field declaration. Re-exported so callers
// don't need to import go-ethereum's apitypes package directly.
type ApiType = apitypes.Type

// SignatureChainIdMain is the chainId Hyperliquid expects in user-signed
// action envelopes. The Python SDK forcibly sets this to 0x66eee (421614)
// regardless of network. Wallets sign over this chainId; the
// `hyperliquidChain` field separately selects Mainnet vs Testnet.
const SignatureChainIdMain = "0x66eee"

// EncodeAction msgpack-encodes an action in canonical (struct-field-declaration)
// order. The action argument should be a typed Go struct with msgpack tags;
// passing a map[string]any will not produce a deterministic hash because Go
// map iteration is randomized.
func EncodeAction(action any) ([]byte, error) {
	return msgpack.Marshal(action)
}

// ActionHash computes the keccak256 hash that gets wrapped into the phantom
// agent's connectionId field. Mirrors signing.py:action_hash exactly.
//
//	hash = keccak256(msgpack(action) || be64(nonce)
//	                || (vault?0x01||addr:0x00)
//	                || (expiresAfter?0x00||be64(expiresAfter):"") )
func ActionHash(action any, vaultAddress *common.Address, nonce uint64, expiresAfter *uint64) ([]byte, error) {
	data, err := EncodeAction(action)
	if err != nil {
		return nil, fmt.Errorf("msgpack action: %w", err)
	}

	var nonceBytes [8]byte
	binary.BigEndian.PutUint64(nonceBytes[:], nonce)
	data = append(data, nonceBytes[:]...)

	if vaultAddress == nil {
		data = append(data, 0x00)
	} else {
		data = append(data, 0x01)
		data = append(data, vaultAddress.Bytes()...)
	}

	if expiresAfter != nil {
		data = append(data, 0x00) // intentional — see package doc comment
		var expBytes [8]byte
		binary.BigEndian.PutUint64(expBytes[:], *expiresAfter)
		data = append(data, expBytes[:]...)
	}

	return crypto.Keccak256(data), nil
}

func phantomSource(isMainnet bool) string {
	if isMainnet {
		return "a"
	}
	return "b"
}

// l1Domain is the EIP-712 domain Hyperliquid uses for L1 action signatures.
// chainId is constant 1337 — NOT Arbitrum's 42161 — even though signatures
// originate from Arbitrum-bound wallets.
func l1Domain() apitypes.TypedDataDomain {
	return apitypes.TypedDataDomain{
		Name:              "Exchange",
		Version:           "1",
		ChainId:           ethmath.NewHexOrDecimal256(1337),
		VerifyingContract: "0x0000000000000000000000000000000000000000",
	}
}

// SignL1Action signs an arbitrary Hyperliquid L1 action. Returns the {r,s,v}
// triple Hyperliquid expects in the request body.
func SignL1Action(
	key *ecdsa.PrivateKey,
	action any,
	vaultAddress *common.Address,
	nonce uint64,
	expiresAfter *uint64,
	isMainnet bool,
) (Signature, error) {
	hash, err := ActionHash(action, vaultAddress, nonce, expiresAfter)
	if err != nil {
		return Signature{}, err
	}
	typedData := apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": []apitypes.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			"Agent": []apitypes.Type{
				{Name: "source", Type: "string"},
				{Name: "connectionId", Type: "bytes32"},
			},
		},
		PrimaryType: "Agent",
		Domain:      l1Domain(),
		Message: apitypes.TypedDataMessage{
			"source":       phantomSource(isMainnet),
			"connectionId": hash,
		},
	}
	return signTyped(key, typedData)
}

// SignUserAction signs a "user-signed" action (transfers, approveAgent,
// approveBuilderFee, withdraw, etc.). The action is signed directly via EIP-712
// using the HyperliquidSignTransaction domain and a per-action primary type +
// payload-types list.
//
// On entry, action MUST already include its `signatureChainId` (hex string)
// and `hyperliquidChain` ("Mainnet" or "Testnet") fields — these participate
// in the typed-data hash. Callers who want the Python-SDK behavior of forcing
// 0x66eee should set those fields explicitly before calling.
func SignUserAction(
	key *ecdsa.PrivateKey,
	action map[string]any,
	primaryType string,
	payloadTypes []ApiType,
) (Signature, error) {
	signatureChainIdHex, _ := action["signatureChainId"].(string)
	if signatureChainIdHex == "" {
		return Signature{}, fmt.Errorf("action.signatureChainId is required for user-signed actions")
	}
	chainIDBig, ok := new(big.Int).SetString(strings.TrimPrefix(signatureChainIdHex, "0x"), 16)
	if !ok {
		return Signature{}, fmt.Errorf("could not parse signatureChainId %q as hex", signatureChainIdHex)
	}

	// Go's apitypes hasher rejects message fields that aren't declared in the
	// type. The Python SDK's eth_account is permissive and signs only the
	// fields it knows. Mimic that here by projecting `action` to the declared
	// payload-types names. `type` and `signatureChainId` deliberately do NOT
	// participate in the typed-data hash — they live alongside.
	//
	// apitypes also requires uint64/uint256 fields to be *big.Int, not native
	// Go integers. Coerce as we project.
	message := apitypes.TypedDataMessage{}
	for _, t := range payloadTypes {
		v, ok := action[t.Name]
		if !ok {
			return Signature{}, fmt.Errorf("action is missing required field %q for primary type %q", t.Name, primaryType)
		}
		switch t.Type {
		case "uint64", "uint256":
			switch x := v.(type) {
			case uint64:
				v = new(big.Int).SetUint64(x)
			case int:
				v = big.NewInt(int64(x))
			case int64:
				v = big.NewInt(x)
			case *big.Int:
				// already correct
			default:
				return Signature{}, fmt.Errorf("field %q: cannot coerce %T to *big.Int", t.Name, v)
			}
		}
		message[t.Name] = v
	}

	typedData := apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": []apitypes.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			primaryType: payloadTypes,
		},
		PrimaryType: primaryType,
		Domain: apitypes.TypedDataDomain{
			Name:              "HyperliquidSignTransaction",
			Version:           "1",
			ChainId:           (*ethmath.HexOrDecimal256)(chainIDBig),
			VerifyingContract: "0x0000000000000000000000000000000000000000",
		},
		Message: message,
	}
	return signTyped(key, typedData)
}

func signTyped(key *ecdsa.PrivateKey, td apitypes.TypedData) (Signature, error) {
	domainSep, err := td.HashStruct("EIP712Domain", td.Domain.Map())
	if err != nil {
		return Signature{}, fmt.Errorf("hash domain: %w", err)
	}
	msgHash, err := td.HashStruct(td.PrimaryType, td.Message)
	if err != nil {
		return Signature{}, fmt.Errorf("hash message: %w", err)
	}
	rawData := []byte{0x19, 0x01}
	rawData = append(rawData, domainSep...)
	rawData = append(rawData, msgHash...)
	digest := crypto.Keccak256(rawData)

	sig, err := crypto.Sign(digest, key)
	if err != nil {
		return Signature{}, fmt.Errorf("ecdsa sign: %w", err)
	}
	if len(sig) != 65 {
		return Signature{}, fmt.Errorf("unexpected signature length %d", len(sig))
	}
	// go-ethereum returns v as 0 or 1; Hyperliquid expects 27 or 28.
	v := sig[64]
	if v < 27 {
		v += 27
	}
	return Signature{
		R: hexutil.Encode(sig[0:32]),
		S: hexutil.Encode(sig[32:64]),
		V: v,
	}, nil
}

// AddressFromKey returns the lowercased 0x address for an ECDSA private key.
func AddressFromKey(key *ecdsa.PrivateKey) string {
	return strings.ToLower(crypto.PubkeyToAddress(key.PublicKey).Hex())
}

// MustParseHexAddress returns a 20-byte address parsed from a 0x... hex string.
// Panics on bad input — callers should pre-validate.
func MustParseHexAddress(hexAddr string) common.Address {
	if !common.IsHexAddress(hexAddr) {
		panic(fmt.Sprintf("not a hex address: %q", hexAddr))
	}
	return common.HexToAddress(hexAddr)
}

// ParseAddress is the non-panicking version.
func ParseAddress(hexAddr string) (common.Address, error) {
	if !common.IsHexAddress(hexAddr) {
		return common.Address{}, fmt.Errorf("not a hex address: %q", hexAddr)
	}
	return common.HexToAddress(hexAddr), nil
}

// HexBytes returns the lowercased 0x-prefixed hex of b.
func HexBytes(b []byte) string {
	return "0x" + hex.EncodeToString(b)
}
