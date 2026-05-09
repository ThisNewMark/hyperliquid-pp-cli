# Builder Fee — full transparency

This CLI ships with a default **builder code**: an address that receives a
small per-trade fee on orders placed through it. Hyperliquid's builder-code
mechanism is consent-based — every fee-bearing order requires you to have
on-chain-approved the builder address with a maximum fee rate. Until you do
that approval, no fee is charged regardless of what flags you pass.

This page documents exactly what's set, how to opt out, and how to verify
on-chain — so you never have to take our word for any of it.

## What's on the wire

Every order placed via `hyperliquid exchange place-order` carries an extra
field on the action:

```json
"builder": { "b": "<builder address>", "f": <fee in tenths of basis points> }
```

The fee unit is **tenths of basis points**: `f=10` means 1 bp = 0.01%.
Hyperliquid caps the fee server-side at 0.1% on perps and 1% on spot.

## The shipping defaults

| Setting | Value | Source |
|---|---|---|
| Default builder address | `0xc8f0cd137e28f717a20f810b46926f92978bbcfa` | `internal/builder/builder.go::DefaultBuilderAddress` |
| Default fee | `10` tenths-of-bps (= 0.01% = 1 bp) | `internal/builder/builder.go::DefaultBuilderFeeBps` |
| Default `maxFeeRate` (for `builder approve`) | `0.01%` | `internal/builder/builder.go::DefaultMaxFeeRate` |

The default fee is **0.01%** — 1 basis point, well below Hyperliquid's 0.1%
perps cap. On a $100 trade, that's one cent.

## How to opt out

Three equivalent ways, any one of them works per-order:

```bash
# Recommended for scripts (most explicit)
hyperliquid exchange place-order --no-builder ...

# Equivalent: zero-address sentinel
hyperliquid exchange place-order --builder 0x0 ...

# Or route fees to your own address
hyperliquid exchange place-order --builder 0xYourAddress --builder-fee-bps 10 ...
```

There is no global "always opt out" flag — the per-order discipline is
intentional, so the choice is always visible at the call site.

## How to verify on-chain

Quick check via the CLI:

```bash
hyperliquid builder status --user <your-main-address>
```

That queries Hyperliquid's `/info {type:approvedBuilders}` and
`/info {type:maxBuilderFee}` endpoints and shows whether you have approved
this CLI's default builder, and at what cap.

To verify without trusting the CLI, curl the public API directly:

```bash
curl -s -X POST https://api.hyperliquid.xyz/info \
  -H 'content-type: application/json' \
  -d '{"type":"approvedBuilders","user":"<your-main-address>"}'
```

You should see the builder address only after you've explicitly approved it.

## How to approve

The setup flow handles this for you (see
[QUICKSTART-CLAUDE-DESKTOP.md](QUICKSTART-CLAUDE-DESKTOP.md) or
[QUICKSTART-TERMINAL.md](QUICKSTART-TERMINAL.md)). If you want to do it
manually:

```bash
hyperliquid builder approve --builder <addr> --max-fee-rate 0.01%
```

This action **must be signed by your main depositing wallet, not an agent**.
Hyperliquid enforces this server-side. To revoke, run:

```bash
hyperliquid builder revoke --builder <addr>
```

…which submits an `approveBuilderFee` action with `maxFeeRate=0%`.
