package sign

import (
	"bytes"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	ethmath "github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

// Tiny canonical action used across tests.
type testAction struct {
	Type     string `msgpack:"type"`
	Greeting string `msgpack:"greeting"`
	N        int    `msgpack:"n"`
}

func TestEncodeAction_Deterministic(t *testing.T) {
	a := testAction{Type: "hello", Greeting: "hi", N: 7}
	b1, err := EncodeAction(a)
	if err != nil {
		t.Fatal(err)
	}
	b2, err := EncodeAction(a)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(b1, b2) {
		t.Fatalf("encoding not deterministic: %x vs %x", b1, b2)
	}
}

func TestActionHash_VaultAndExpiresAfter(t *testing.T) {
	a := testAction{Type: "x", Greeting: "y", N: 1}

	// All four shapes should produce distinct hashes.
	noVaultNoExp, _ := ActionHash(a, nil, 1, nil)

	vault := common.HexToAddress("0x0000000000000000000000000000000000000001")
	withVault, _ := ActionHash(a, &vault, 1, nil)

	exp := uint64(2)
	noVaultWithExp, _ := ActionHash(a, nil, 1, &exp)

	withBoth, _ := ActionHash(a, &vault, 1, &exp)

	all := [][]byte{noVaultNoExp, withVault, noVaultWithExp, withBoth}
	for i := range all {
		for j := i + 1; j < len(all); j++ {
			if bytes.Equal(all[i], all[j]) {
				t.Fatalf("hash %d == hash %d (vault/expires combinations should differ)", i, j)
			}
		}
	}

	// Hashes are 32 bytes (keccak256).
	for i, h := range all {
		if len(h) != 32 {
			t.Errorf("hash[%d] length = %d, want 32", i, len(h))
		}
	}
}

func TestSignL1Action_Recovers(t *testing.T) {
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	addr := strings.ToLower(crypto.PubkeyToAddress(key.PublicKey).Hex())

	a := testAction{Type: "x", Greeting: "y", N: 42}
	sig, err := SignL1Action(key, a, nil, 12345, nil, true)
	if err != nil {
		t.Fatalf("SignL1Action: %v", err)
	}
	if sig.V != 27 && sig.V != 28 {
		t.Errorf("v = %d, want 27 or 28", sig.V)
	}

	// Reconstruct the typed-data digest the same way SignL1Action did, then
	// recover the signer's pubkey from the signature.
	hash, _ := ActionHash(a, nil, 12345, nil)
	td := apitypes.TypedData{
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
			"source":       "a",
			"connectionId": hash,
		},
	}
	domainSep, err := td.HashStruct("EIP712Domain", td.Domain.Map())
	if err != nil {
		t.Fatal(err)
	}
	msgHash, err := td.HashStruct(td.PrimaryType, td.Message)
	if err != nil {
		t.Fatal(err)
	}
	raw := append([]byte{0x19, 0x01}, domainSep...)
	raw = append(raw, msgHash...)
	digest := crypto.Keccak256(raw)

	// Reassemble {r,s,v} bytes (with v normalized back to 0/1 for ecrecover).
	rBytes := common.FromHex(sig.R)
	sBytes := common.FromHex(sig.S)
	v := sig.V
	if v >= 27 {
		v -= 27
	}
	rsv := append(append(rBytes, sBytes...), v)

	pub, err := crypto.SigToPub(digest, rsv)
	if err != nil {
		t.Fatalf("SigToPub: %v", err)
	}
	recovered := strings.ToLower(crypto.PubkeyToAddress(*pub).Hex())
	if recovered != addr {
		t.Fatalf("recovered %s, want %s", recovered, addr)
	}
}

func TestSignUserAction_ApproveBuilderFee(t *testing.T) {
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	addr := strings.ToLower(crypto.PubkeyToAddress(key.PublicKey).Hex())

	action := map[string]any{
		"type":             "approveBuilderFee",
		"hyperliquidChain": "Mainnet",
		"signatureChainId": "0x66eee",
		"maxFeeRate":       "0.01%",
		"builder":          "0xc8f0cd137e28f717a20f810b46926f92978bbcfa",
		"nonce":            uint64(1700000000000),
	}
	payloadTypes := []apitypes.Type{
		{Name: "hyperliquidChain", Type: "string"},
		{Name: "maxFeeRate", Type: "string"},
		{Name: "builder", Type: "address"},
		{Name: "nonce", Type: "uint64"},
	}
	sig, err := SignUserAction(key, action, "HyperliquidTransaction:ApproveBuilderFee", payloadTypes)
	if err != nil {
		t.Fatalf("SignUserAction: %v", err)
	}
	if sig.V != 27 && sig.V != 28 {
		t.Errorf("v = %d, want 27 or 28", sig.V)
	}
	if !strings.HasPrefix(sig.R, "0x") || !strings.HasPrefix(sig.S, "0x") {
		t.Errorf("r/s missing 0x prefix: r=%s s=%s", sig.R, sig.S)
	}

	// Sanity: re-derive the digest exactly like SignUserAction does and
	// confirm we recover the signer.
	chainID, _ := new(big.Int).SetString("66eee", 16)
	td := apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": []apitypes.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			"HyperliquidTransaction:ApproveBuilderFee": payloadTypes,
		},
		PrimaryType: "HyperliquidTransaction:ApproveBuilderFee",
		Domain: apitypes.TypedDataDomain{
			Name:              "HyperliquidSignTransaction",
			Version:           "1",
			ChainId:           (*ethmath.HexOrDecimal256)(chainID),
			VerifyingContract: "0x0000000000000000000000000000000000000000",
		},
		Message: apitypes.TypedDataMessage{
			"hyperliquidChain": action["hyperliquidChain"],
			"maxFeeRate":       action["maxFeeRate"],
			"builder":          action["builder"],
			"nonce":            new(big.Int).SetUint64(action["nonce"].(uint64)),
		},
	}
	domainSep, err := td.HashStruct("EIP712Domain", td.Domain.Map())
	if err != nil {
		t.Fatal(err)
	}
	msgHash, err := td.HashStruct(td.PrimaryType, td.Message)
	if err != nil {
		t.Fatal(err)
	}
	raw := append([]byte{0x19, 0x01}, domainSep...)
	raw = append(raw, msgHash...)
	digest := crypto.Keccak256(raw)

	rBytes := common.FromHex(sig.R)
	sBytes := common.FromHex(sig.S)
	v := sig.V
	if v >= 27 {
		v -= 27
	}
	rsv := append(append(rBytes, sBytes...), v)
	pub, err := crypto.SigToPub(digest, rsv)
	if err != nil {
		t.Fatalf("SigToPub: %v", err)
	}
	if got := strings.ToLower(crypto.PubkeyToAddress(*pub).Hex()); got != addr {
		t.Fatalf("recovered %s, want %s", got, addr)
	}
}
