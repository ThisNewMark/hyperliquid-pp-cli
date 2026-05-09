package sign

import "testing"

// Test cases mirror the Python SDK's float_to_wire output (Decimal.normalize)
// so our msgpack matches HL's canonical form. If this drifts, place-order
// will fail with "User or API Wallet does not exist."
func TestNormalizeDecimal(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"5", "5"},
		{"5.0", "5"},
		{"5.00000000", "5"},
		{"0.70", "0.7"},
		{"0.001", "0.001"},
		{"0.001000", "0.001"},
		{"45.22", "45.22"},
		{"45.220", "45.22"},
		{"100.00", "100"},
		{"100", "100"},
		{"-1.50", "-1.5"},
		{"-0", "0"},
		{"-0.0", "0"},
		{"-0.00", "0"},
		{"30000", "30000"},
		{"0", "0"},
		{"0.0", "0"},
	}
	for _, c := range cases {
		got := NormalizeDecimal(c.in)
		if got != c.want {
			t.Errorf("NormalizeDecimal(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestNewPlaceOrderAction_NormalizesPriceAndSize(t *testing.T) {
	orders := []OrderWire{{
		Asset: 159, IsBuy: true,
		LimitPx: "5.0",  // should become "5"
		Size:    "6.00", // should become "6"
		OrderType: OrderTypeWire{Limit: &LimitOrderType{Tif: "Alo"}},
	}}
	action := NewPlaceOrderAction(orders, "na", nil)
	if action.Orders[0].LimitPx != "5" {
		t.Errorf("LimitPx not normalized: got %q, want %q", action.Orders[0].LimitPx, "5")
	}
	if action.Orders[0].Size != "6" {
		t.Errorf("Size not normalized: got %q, want %q", action.Orders[0].Size, "6")
	}
}
