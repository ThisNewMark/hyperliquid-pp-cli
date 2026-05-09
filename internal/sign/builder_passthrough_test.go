package sign

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
)

// TestBuilderFieldPassthrough is the load-bearing test for this CLI's whole
// monetization angle: it asserts that when a PlaceOrderAction carries a
// Builder field, the field is (a) present in the msgpack-encoded action, and
// (b) actually changes the action hash that gets signed.
//
// If Builder were silently dropped at any step, this test would fail. Any
// future SDK swap or refactor MUST keep this test green.
func TestBuilderFieldPassthrough(t *testing.T) {
	addr := "0xc8f0cd137e28f717a20f810b46926f92978bbcfa"
	feeBps := 1

	orders := []OrderWire{
		{
			Asset:     0,
			IsBuy:     true,
			LimitPx:   "100",
			Size:      "0.1",
			OrderType: OrderTypeWire{Limit: &LimitOrderType{Tif: "Gtc"}},
		},
	}

	noBuilder := NewPlaceOrderAction(orders, "na", nil)
	withBuilder := NewPlaceOrderAction(orders, "na", &BuilderInfo{B: addr, F: feeBps})

	encNo, err := EncodeAction(noBuilder)
	if err != nil {
		t.Fatal(err)
	}
	encYes, err := EncodeAction(withBuilder)
	if err != nil {
		t.Fatal(err)
	}

	// (a) Builder field must NOT appear when nil — it's `omitempty`.
	if bytes.Contains(encNo, []byte("builder")) {
		t.Errorf("nil-builder action contains 'builder' key in msgpack: %x", encNo)
	}
	if bytes.Contains(encNo, []byte(addr)) {
		t.Errorf("nil-builder action contains the address bytes: %x", encNo)
	}

	// (b) Builder field MUST appear when set.
	if !bytes.Contains(encYes, []byte("builder")) {
		t.Errorf("builder action does NOT contain 'builder' key in msgpack: %x", encYes)
	}
	if !bytes.Contains(encYes, []byte(addr)) {
		t.Errorf("builder action does NOT contain the address bytes: %x", encYes)
	}

	// (c) The hashes that get signed must differ.
	hashNo, _ := ActionHash(noBuilder, nil, 1717000000000, nil)
	hashYes, _ := ActionHash(withBuilder, nil, 1717000000000, nil)
	if bytes.Equal(hashNo, hashYes) {
		t.Errorf("action hashes are equal — builder field is being silently dropped before keccak")
	}

	// (d) Two different builder fees MUST also produce different hashes —
	// otherwise our `f` field is being ignored.
	withBuilderHigherFee := NewPlaceOrderAction(orders, "na", &BuilderInfo{B: addr, F: 5})
	hashHigh, _ := ActionHash(withBuilderHigherFee, nil, 1717000000000, nil)
	if bytes.Equal(hashYes, hashHigh) {
		t.Errorf("action hashes equal across f=1 and f=5 — `f` field is being dropped")
	}

	// (e) End-to-end: signing the with-builder action produces a signature
	// different from signing the no-builder one. (Different digests sign to
	// different (r,s) almost surely.)
	key, _ := crypto.GenerateKey()
	sigNo, err := SignL1Action(key, noBuilder, nil, 1717000000000, nil, true)
	if err != nil {
		t.Fatal(err)
	}
	sigYes, err := SignL1Action(key, withBuilder, nil, 1717000000000, nil, true)
	if err != nil {
		t.Fatal(err)
	}
	if sigNo.R == sigYes.R && sigNo.S == sigYes.S {
		t.Errorf("identical signatures across builder/no-builder — pipeline is broken")
	}
}

// TestPlaceOrderActionFieldOrder asserts msgpack writes the keys in the
// order Hyperliquid's docs declare: type, orders, grouping, builder. If a
// future struct refactor reorders fields, the action hash silently changes
// and Hyperliquid will reject orders with "Invalid signature".
func TestPlaceOrderActionFieldOrder(t *testing.T) {
	a := NewPlaceOrderAction(
		[]OrderWire{{Asset: 0, IsBuy: true, LimitPx: "100", Size: "1", OrderType: OrderTypeWire{Limit: &LimitOrderType{Tif: "Gtc"}}}},
		"na",
		&BuilderInfo{B: "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", F: 7},
	)
	enc, err := EncodeAction(a)
	if err != nil {
		t.Fatal(err)
	}
	encStr := string(enc)
	// Confirm the keys appear in the right relative order. msgpack keys are
	// written as raw strings prefixed by their length byte; substring search
	// is enough for this assertion.
	wantOrder := []string{"type", "orders", "grouping", "builder"}
	prev := -1
	for _, k := range wantOrder {
		idx := strings.Index(encStr, k)
		if idx < 0 {
			t.Fatalf("key %q not found in msgpack output", k)
		}
		if idx <= prev {
			t.Fatalf("key %q appeared at index %d, before previous key (idx %d): wrong field order", k, idx, prev)
		}
		prev = idx
	}
}
