package sign

import "strings"

// Hyperliquid action structs in their canonical wire shape. msgpack field
// order matters for the action hash — defined here in the order the official
// Python SDK uses (struct definition order = msgpack key order in v5).
//
// Adding a field here? Make sure to:
//  1. put it in the order that matches the Python SDK / the docs;
//  2. mark it `,omitempty` only if Hyperliquid actually permits it absent
//     (otherwise the hash you sign will not match what Hyperliquid expects);
//  3. for any decimal-string field (price, size, trigger price), call
//     NormalizeDecimal on the user input BEFORE constructing the wire struct
//     — otherwise "5.0" vs "5" or "0.70" vs "0.7" will produce a different
//     msgpack and HL will reject the signature with "User or API Wallet
//     does not exist".

// BuilderInfo is the per-order builder-code field. `f` is in tenths of basis
// points (10 = 1bp = 0.01%). Hyperliquid caps builder fees at 0.1% perps,
// 1% spot. Pre-approve the builder address with ApproveBuilderFee first.
type BuilderInfo struct {
	B string `msgpack:"b" json:"b"`
	F int    `msgpack:"f" json:"f"`
}

// NormalizeDecimal strips trailing zeros from a decimal-string price/size to
// match Hyperliquid's canonical wire format. Mirrors the Python SDK's
// float_to_wire(...) behavior, which calls Decimal(x).normalize() before
// emitting the value into the action.
//
// Without this normalization, two strings that look numerically equal (e.g.
// "5.0" vs "5", or "0.70" vs "0.7") produce different msgpack output, which
// produces a different action_hash, which produces a different recovered
// signer on HL's side. HL then returns "User or API Wallet does not exist"
// because the recovered (garbage) address isn't a registered agent.
//
//	NormalizeDecimal("5")          = "5"
//	NormalizeDecimal("5.0")        = "5"
//	NormalizeDecimal("5.00000000") = "5"
//	NormalizeDecimal("0.70")       = "0.7"
//	NormalizeDecimal("0.001")      = "0.001"
//	NormalizeDecimal("45.22")      = "45.22"
//	NormalizeDecimal("100.00")     = "100"
//	NormalizeDecimal("-0")         = "0"
//	NormalizeDecimal("-0.0")       = "0"
//	NormalizeDecimal("-1.50")      = "-1.5"
func NormalizeDecimal(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	if !strings.ContainsRune(s, '.') {
		// Integers normalize to themselves. "-0" handled below.
		if s == "-0" {
			return "0"
		}
		return s
	}
	// Has a decimal point. Strip trailing zeros, then a dangling ".".
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	if s == "" || s == "-" {
		return "0"
	}
	if s == "-0" {
		return "0"
	}
	return s
}

// normalizeOrderWire applies NormalizeDecimal to the price/size/trigger
// fields of an OrderWire so the msgpack output matches HL's canonical form.
func normalizeOrderWire(o OrderWire) OrderWire {
	o.LimitPx = NormalizeDecimal(o.LimitPx)
	o.Size = NormalizeDecimal(o.Size)
	if o.OrderType.Trigger != nil {
		t := *o.OrderType.Trigger
		t.TriggerPx = NormalizeDecimal(t.TriggerPx)
		o.OrderType.Trigger = &t
	}
	return o
}

// ---------------------------------------------------------------------------
// L1 actions (signed via phantom-agent, msgpack-hashed)
// ---------------------------------------------------------------------------

type LimitOrderType struct {
	Tif string `msgpack:"tif" json:"tif"`
}

type TriggerOrderType struct {
	IsMarket  bool   `msgpack:"isMarket" json:"isMarket"`
	TriggerPx string `msgpack:"triggerPx" json:"triggerPx"`
	Tpsl      string `msgpack:"tpsl" json:"tpsl"`
}

type OrderTypeWire struct {
	Limit   *LimitOrderType   `msgpack:"limit,omitempty" json:"limit,omitempty"`
	Trigger *TriggerOrderType `msgpack:"trigger,omitempty" json:"trigger,omitempty"`
}

// OrderWire is one element of PlaceOrderAction.Orders. Matches the Hyperliquid
// /exchange "order" action shape exactly.
type OrderWire struct {
	Asset      int           `msgpack:"a" json:"a"`
	IsBuy      bool          `msgpack:"b" json:"b"`
	LimitPx    string        `msgpack:"p" json:"p"`
	Size       string        `msgpack:"s" json:"s"`
	ReduceOnly bool          `msgpack:"r" json:"r"`
	OrderType  OrderTypeWire `msgpack:"t" json:"t"`
	Cloid      string        `msgpack:"c,omitempty" json:"c,omitempty"`
}

// PlaceOrderAction is the body of the "order" action.
//
// CRITICAL: Builder must be a struct field declared AFTER Grouping so the
// msgpack key order matches what Hyperliquid expects. If Builder is nil
// (`*BuilderInfo`), it is omitted entirely from the msgpack output via the
// `omitempty` tag — that's the opt-out path.
type PlaceOrderAction struct {
	Type     string       `msgpack:"type" json:"type"`
	Orders   []OrderWire  `msgpack:"orders" json:"orders"`
	Grouping string       `msgpack:"grouping" json:"grouping"`
	Builder  *BuilderInfo `msgpack:"builder,omitempty" json:"builder,omitempty"`
}

// NewPlaceOrderAction builds a properly-typed action ready for signing.
// Pass `builder` as nil to opt out of builder fees on this batch.
//
// Normalizes each order's price/size strings to HL canonical form (strips
// trailing zeros) so the signed action_hash matches what HL recomputes.
func NewPlaceOrderAction(orders []OrderWire, grouping string, builder *BuilderInfo) *PlaceOrderAction {
	if grouping == "" {
		grouping = "na"
	}
	out := make([]OrderWire, len(orders))
	for i, o := range orders {
		out[i] = normalizeOrderWire(o)
	}
	return &PlaceOrderAction{
		Type:     "order",
		Orders:   out,
		Grouping: grouping,
		Builder:  builder,
	}
}

type CancelOidWire struct {
	Asset int `msgpack:"a" json:"a"`
	Oid   int `msgpack:"o" json:"o"`
}

type CancelOidAction struct {
	Type    string          `msgpack:"type" json:"type"`
	Cancels []CancelOidWire `msgpack:"cancels" json:"cancels"`
}

func NewCancelOidAction(cancels []CancelOidWire) *CancelOidAction {
	return &CancelOidAction{Type: "cancel", Cancels: cancels}
}

type CancelCloidWire struct {
	Asset int    `msgpack:"asset" json:"asset"`
	Cloid string `msgpack:"cloid" json:"cloid"`
}

type CancelCloidAction struct {
	Type    string            `msgpack:"type" json:"type"`
	Cancels []CancelCloidWire `msgpack:"cancels" json:"cancels"`
}

func NewCancelCloidAction(cancels []CancelCloidWire) *CancelCloidAction {
	return &CancelCloidAction{Type: "cancelByCloid", Cancels: cancels}
}

type ModifyAction struct {
	Type  string    `msgpack:"type" json:"type"`
	Oid   int       `msgpack:"oid" json:"oid"`
	Order OrderWire `msgpack:"order" json:"order"`
}

func NewModifyAction(oid int, order OrderWire) *ModifyAction {
	return &ModifyAction{Type: "modify", Oid: oid, Order: normalizeOrderWire(order)}
}

type ModifyWire struct {
	Oid   int       `msgpack:"oid" json:"oid"`
	Order OrderWire `msgpack:"order" json:"order"`
}

type BatchModifyAction struct {
	Type     string       `msgpack:"type" json:"type"`
	Modifies []ModifyWire `msgpack:"modifies" json:"modifies"`
}

func NewBatchModifyAction(modifies []ModifyWire) *BatchModifyAction {
	out := make([]ModifyWire, len(modifies))
	for i, m := range modifies {
		out[i] = ModifyWire{Oid: m.Oid, Order: normalizeOrderWire(m.Order)}
	}
	return &BatchModifyAction{Type: "batchModify", Modifies: out}
}

type UpdateLeverageAction struct {
	Type     string `msgpack:"type" json:"type"`
	Asset    int    `msgpack:"asset" json:"asset"`
	IsCross  bool   `msgpack:"isCross" json:"isCross"`
	Leverage int    `msgpack:"leverage" json:"leverage"`
}

func NewUpdateLeverageAction(asset int, isCross bool, leverage int) *UpdateLeverageAction {
	return &UpdateLeverageAction{Type: "updateLeverage", Asset: asset, IsCross: isCross, Leverage: leverage}
}

type UpdateIsolatedMarginAction struct {
	Type  string `msgpack:"type" json:"type"`
	Asset int    `msgpack:"asset" json:"asset"`
	IsBuy bool   `msgpack:"isBuy" json:"isBuy"`
	Ntli  int    `msgpack:"ntli" json:"ntli"`
}

func NewUpdateIsolatedMarginAction(asset int, isBuy bool, ntli int) *UpdateIsolatedMarginAction {
	return &UpdateIsolatedMarginAction{Type: "updateIsolatedMargin", Asset: asset, IsBuy: isBuy, Ntli: ntli}
}

type ScheduleCancelAction struct {
	Type string `msgpack:"type" json:"type"`
	Time *int64 `msgpack:"time,omitempty" json:"time,omitempty"`
}

func NewScheduleCancelAction(time *int64) *ScheduleCancelAction {
	return &ScheduleCancelAction{Type: "scheduleCancel", Time: time}
}

type TwapWire struct {
	Asset      int    `msgpack:"a" json:"a"`
	IsBuy      bool   `msgpack:"b" json:"b"`
	Size       string `msgpack:"s" json:"s"`
	ReduceOnly bool   `msgpack:"r" json:"r"`
	Minutes    int    `msgpack:"m" json:"m"`
	Randomize  bool   `msgpack:"t" json:"t"`
}

type TwapOrderAction struct {
	Type string   `msgpack:"type" json:"type"`
	Twap TwapWire `msgpack:"twap" json:"twap"`
}

func NewTwapOrderAction(t TwapWire) *TwapOrderAction {
	t.Size = NormalizeDecimal(t.Size)
	return &TwapOrderAction{Type: "twapOrder", Twap: t}
}

type TwapCancelAction struct {
	Type   string `msgpack:"type" json:"type"`
	Asset  int    `msgpack:"a" json:"a"`
	TwapId int    `msgpack:"t" json:"t"`
}

func NewTwapCancelAction(asset, twapID int) *TwapCancelAction {
	return &TwapCancelAction{Type: "twapCancel", Asset: asset, TwapId: twapID}
}

// VaultTransferAction is an L1 (msgpack-hashed) action despite involving
// USD movement — see Hyperliquid docs.
type VaultTransferAction struct {
	Type         string  `msgpack:"type" json:"type"`
	VaultAddress string  `msgpack:"vaultAddress" json:"vaultAddress"`
	IsDeposit    bool    `msgpack:"isDeposit" json:"isDeposit"`
	Usd          float64 `msgpack:"usd" json:"usd"`
}

func NewVaultTransferAction(vault string, isDeposit bool, usd float64) *VaultTransferAction {
	return &VaultTransferAction{Type: "vaultTransfer", VaultAddress: vault, IsDeposit: isDeposit, Usd: usd}
}

// ---------------------------------------------------------------------------
// User-signed actions: payload-type definitions only. The action body itself
// is built as a map[string]any in the calling code so we can use the same
// map for both signing and posting.
// ---------------------------------------------------------------------------

// Convenience constructors for the actions this CLI cares about. Each returns
// (primaryType, payloadTypes) ready to feed into SignUserAction.

func ApproveBuilderFeeSpec() (string, []ApiType) {
	return "HyperliquidTransaction:ApproveBuilderFee", []ApiType{
		{Name: "hyperliquidChain", Type: "string"},
		{Name: "maxFeeRate", Type: "string"},
		{Name: "builder", Type: "address"},
		{Name: "nonce", Type: "uint64"},
	}
}

func ApproveAgentSpec() (string, []ApiType) {
	return "HyperliquidTransaction:ApproveAgent", []ApiType{
		{Name: "hyperliquidChain", Type: "string"},
		{Name: "agentAddress", Type: "address"},
		{Name: "agentName", Type: "string"},
		{Name: "nonce", Type: "uint64"},
	}
}

func WithdrawSpec() (string, []ApiType) {
	return "HyperliquidTransaction:Withdraw", []ApiType{
		{Name: "hyperliquidChain", Type: "string"},
		{Name: "destination", Type: "string"},
		{Name: "amount", Type: "string"},
		{Name: "time", Type: "uint64"},
	}
}

func UsdSendSpec() (string, []ApiType) {
	return "HyperliquidTransaction:UsdSend", []ApiType{
		{Name: "hyperliquidChain", Type: "string"},
		{Name: "destination", Type: "string"},
		{Name: "amount", Type: "string"},
		{Name: "time", Type: "uint64"},
	}
}

func SpotSendSpec() (string, []ApiType) {
	return "HyperliquidTransaction:SpotSend", []ApiType{
		{Name: "hyperliquidChain", Type: "string"},
		{Name: "destination", Type: "string"},
		{Name: "token", Type: "string"},
		{Name: "amount", Type: "string"},
		{Name: "time", Type: "uint64"},
	}
}

func UsdClassTransferSpec() (string, []ApiType) {
	return "HyperliquidTransaction:UsdClassTransfer", []ApiType{
		{Name: "hyperliquidChain", Type: "string"},
		{Name: "amount", Type: "string"},
		{Name: "toPerp", Type: "bool"},
		{Name: "nonce", Type: "uint64"},
	}
}

func TokenDelegateSpec() (string, []ApiType) {
	return "HyperliquidTransaction:TokenDelegate", []ApiType{
		{Name: "hyperliquidChain", Type: "string"},
		{Name: "validator", Type: "address"},
		{Name: "wei", Type: "uint64"},
		{Name: "isUndelegate", Type: "bool"},
		{Name: "nonce", Type: "uint64"},
	}
}

func CDepositSpec() (string, []ApiType) {
	return "HyperliquidTransaction:CDeposit", []ApiType{
		{Name: "hyperliquidChain", Type: "string"},
		{Name: "wei", Type: "uint64"},
		{Name: "nonce", Type: "uint64"},
	}
}

func CWithdrawSpec() (string, []ApiType) {
	return "HyperliquidTransaction:CWithdraw", []ApiType{
		{Name: "hyperliquidChain", Type: "string"},
		{Name: "wei", Type: "uint64"},
		{Name: "nonce", Type: "uint64"},
	}
}
