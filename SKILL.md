---
name: pp-hyperliquid
description: "Printing Press CLI for Hyperliquid. Hyperliquid is an on-chain order-book L1 perpetuals + spot exchange. The public HTTP API exposes two physical..."
author: "Mark S"
license: "Apache-2.0"
argument-hint: "<command> [args] | install cli|mcp"
allowed-tools: "Read Bash"
metadata:
  openclaw:
    requires:
      bins:
        - hyperliquid-pp-cli
    install:
      - kind: go
        bins: [hyperliquid-pp-cli]
        module: github.com/mvanhorn/printing-press-library/library/other/hyperliquid/cmd/hyperliquid-pp-cli
---

# Hyperliquid — Printing Press CLI

## Prerequisites: Install the CLI

This skill drives the `hyperliquid-pp-cli` binary. **You must verify the CLI is installed before invoking any command from this skill.** If it is missing, install it first:

1. Install via the Printing Press installer:
   ```bash
   npx -y @mvanhorn/printing-press install hyperliquid --cli-only
   ```
2. Verify: `hyperliquid-pp-cli --version`
3. Ensure `$GOPATH/bin` (or `$HOME/go/bin`) is on `$PATH`.

If the `npx` install fails (no Node, offline, etc.), fall back to a direct Go install (requires Go 1.26.3 or newer):

```bash
go install github.com/mvanhorn/printing-press-library/library/other/hyperliquid/cmd/hyperliquid-pp-cli@latest
```

If `--version` reports "command not found" after install, the install step did not put the binary on `$PATH`. Do not proceed with skill commands until verification succeeds.

Hyperliquid is an on-chain order-book L1 perpetuals + spot exchange. The
public HTTP API exposes two physical endpoints — `/info` (read) and
`/exchange` (write) — that multiplex many logical actions via a `type`
discriminator in the request body.

For CLI ergonomics, this spec models each discriminator value as its own
path. The generated client wraps each call in the appropriate envelope
targeting `/info` or `/exchange` against the configured network base URL.

## HTTP Transport

This CLI uses Chrome-compatible HTTP transport for browser-facing endpoints. It does not require a resident browser process for normal API calls.

## Command Reference

**exchange** — Manage exchange

- `hyperliquid-pp-cli exchange approve-agent` — Approve an agent (API) wallet for trading
- `hyperliquid-pp-cli exchange approve-builder-fee` — MUST be signed by the user's main wallet, not an agent. Max fee rate is a human-readable percentage string like...
- `hyperliquid-pp-cli exchange batch-modify-orders` — Modify multiple orders atomically
- `hyperliquid-pp-cli exchange cancel-orders` — Cancel orders by oid
- `hyperliquid-pp-cli exchange cancel-orders-by-cloid` — Cancel orders by client order id
- `hyperliquid-pp-cli exchange cancel-twap-order` — Cancel a TWAP order
- `hyperliquid-pp-cli exchange class-transfer` — Move USDC between perp and spot
- `hyperliquid-pp-cli exchange delegate-token` — Delegate (or undelegate) staked HYPE to a validator
- `hyperliquid-pp-cli exchange modify-order` — Modify a single order
- `hyperliquid-pp-cli exchange place-order` — Place one or more orders. Optionally attaches a builder code.
- `hyperliquid-pp-cli exchange place-twap-order` — Place a TWAP order
- `hyperliquid-pp-cli exchange schedule-cancel` — Dead-man-switch — cancel all orders at a future timestamp
- `hyperliquid-pp-cli exchange send-spot` — Transfer a spot asset
- `hyperliquid-pp-cli exchange send-usd` — Transfer USDC on the perps account
- `hyperliquid-pp-cli exchange stake-deposit` — Stake HYPE into the staking layer
- `hyperliquid-pp-cli exchange stake-withdraw` — Begin a 7-day unstake from the staking layer
- `hyperliquid-pp-cli exchange update-isolated-margin` — Add/remove isolated margin on an asset
- `hyperliquid-pp-cli exchange update-leverage` — Update cross/isolated leverage for an asset
- `hyperliquid-pp-cli exchange vault-transfer` — Deposit to or withdraw from a vault
- `hyperliquid-pp-cli exchange withdraw` — Withdraw USDC to L1 (~5min, $1 fee)

**info** — Read-only queries against /info

- `hyperliquid-pp-cli info get-all-mids` — All mid prices
- `hyperliquid-pp-cli info get-approved-builders` — List of builder addresses this user has approved
- `hyperliquid-pp-cli info get-candle-snapshot` — Candles for a coin and interval
- `hyperliquid-pp-cli info get-clearinghouse-state` — Perp positions, margin summary, withdrawable
- `hyperliquid-pp-cli info get-frontend-open-orders` — User's open orders (frontend form, includes triggers)
- `hyperliquid-pp-cli info get-funding-history` — Historical funding rates for a coin
- `hyperliquid-pp-cli info get-historical-orders` — Recent historical orders
- `hyperliquid-pp-cli info get-l2-book` — L2 order book for a coin
- `hyperliquid-pp-cli info get-max-builder-fee` — Max builder fee in tenths of bps that this user has approved for this builder
- `hyperliquid-pp-cli info get-open-orders` — User's open orders (compact form)
- `hyperliquid-pp-cli info get-order-status` — Status for a single order by oid or cloid
- `hyperliquid-pp-cli info get-perp-meta` — Get perpetuals universe and margin tables
- `hyperliquid-pp-cli info get-perp-meta-and-asset-ctxs` — Perp metadata plus per-asset mark/funding context
- `hyperliquid-pp-cli info get-perps-at-open-interest-cap` — Coins currently at OI cap
- `hyperliquid-pp-cli info get-portfolio` — Account value and PnL history (day/week/month/allTime)
- `hyperliquid-pp-cli info get-predicted-fundings` — Predicted next funding rates across venues
- `hyperliquid-pp-cli info get-referral` — Referral state including builder rewards
- `hyperliquid-pp-cli info get-spot-clearinghouse-state` — Spot balances
- `hyperliquid-pp-cli info get-spot-meta` — Spot universe metadata
- `hyperliquid-pp-cli info get-spot-meta-and-asset-ctxs` — Spot metadata plus per-asset context
- `hyperliquid-pp-cli info get-sub-accounts` — Sub-accounts owned by this address
- `hyperliquid-pp-cli info get-user-fees` — User fee schedule and current rates
- `hyperliquid-pp-cli info get-user-fills` — User fills (most recent ~2000)
- `hyperliquid-pp-cli info get-user-fills-by-time` — User fills in a time range
- `hyperliquid-pp-cli info get-user-funding` — User funding payment history
- `hyperliquid-pp-cli info get-user-non-funding-ledger-updates` — Deposits, withdrawals, transfers (non-funding ledger)
- `hyperliquid-pp-cli info get-user-role` — Identify whether an address is user/agent/vault/subAccount/missing
- `hyperliquid-pp-cli info get-user-vault-equities` — User's equity in each vault
- `hyperliquid-pp-cli info get-vault-details` — Details for a single vault


### Finding the right command

When you know what you want to do but not which command does it, ask the CLI directly:

```bash
hyperliquid-pp-cli which "<capability in your own words>"
```

`which` resolves a natural-language capability query to the best matching command from this CLI's curated feature index. Exit code `0` means at least one match; exit code `2` means no confident match — fall back to `--help` or use a narrower query.

## Auth Setup

No authentication required.

Run `hyperliquid-pp-cli doctor` to verify setup.

## Agent Mode

Add `--agent` to any command. Expands to: `--json --compact --no-input --no-color --yes`.

- **Pipeable** — JSON on stdout, errors on stderr
- **Filterable** — `--select` keeps a subset of fields. Dotted paths descend into nested structures; arrays traverse element-wise. Critical for keeping context small on verbose APIs:

  ```bash
  hyperliquid-pp-cli exchange approve-agent --agent-address example-value --agent --select id,name,status
  ```
- **Previewable** — `--dry-run` shows the request without sending
- **Offline-friendly** — sync/search commands can use the local SQLite store when available
- **Non-interactive** — never prompts, every input is a flag
- **Explicit retries** — use `--idempotent` only when an already-existing create should count as success

### Response envelope

Commands that read from the local store or the API wrap output in a provenance envelope:

```json
{
  "meta": {"source": "live" | "local", "synced_at": "...", "reason": "..."},
  "results": <data>
}
```

Parse `.results` for data and `.meta.source` to know whether it's live or local. A human-readable `N results (live)` summary is printed to stderr only when stdout is a terminal — piped/agent consumers get pure JSON on stdout.

## Agent Feedback

When you (or the agent) notice something off about this CLI, record it:

```
hyperliquid-pp-cli feedback "the --since flag is inclusive but docs say exclusive"
hyperliquid-pp-cli feedback --stdin < notes.txt
hyperliquid-pp-cli feedback list --json --limit 10
```

Entries are stored locally at `~/.hyperliquid-pp-cli/feedback.jsonl`. They are never POSTed unless `HYPERLIQUID_FEEDBACK_ENDPOINT` is set AND either `--send` is passed or `HYPERLIQUID_FEEDBACK_AUTO_SEND=true`. Default behavior is local-only.

Write what *surprised* you, not a bug report. Short, specific, one line: that is the part that compounds.

## Output Delivery

Every command accepts `--deliver <sink>`. The output goes to the named sink in addition to (or instead of) stdout, so agents can route command results without hand-piping. Three sinks are supported:

| Sink | Effect |
|------|--------|
| `stdout` | Default; write to stdout only |
| `file:<path>` | Atomically write output to `<path>` (tmp + rename) |
| `webhook:<url>` | POST the output body to the URL (`application/json` or `application/x-ndjson` when `--compact`) |

Unknown schemes are refused with a structured error naming the supported set. Webhook failures return non-zero and log the URL + HTTP status on stderr.

## Named Profiles

A profile is a saved set of flag values, reused across invocations. Use it when a scheduled agent calls the same command every run with the same configuration - HeyGen's "Beacon" pattern.

```
hyperliquid-pp-cli profile save briefing --json
hyperliquid-pp-cli --profile briefing exchange approve-agent --agent-address example-value
hyperliquid-pp-cli profile list --json
hyperliquid-pp-cli profile show briefing
hyperliquid-pp-cli profile delete briefing --yes
```

Explicit flags always win over profile values; profile values win over defaults. `agent-context` lists all available profiles under `available_profiles` so introspecting agents discover them at runtime.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 2 | Usage error (wrong arguments) |
| 3 | Resource not found |
| 5 | API error (upstream issue) |
| 7 | Rate limited (wait and retry) |
| 10 | Config error |

## Argument Parsing

Parse `$ARGUMENTS`:

1. **Empty, `help`, or `--help`** → show `hyperliquid-pp-cli --help` output
2. **Starts with `install`** → ends with `mcp` → MCP installation; otherwise → see Prerequisites above
3. **Anything else** → Direct Use (execute as CLI command with `--agent`)

## MCP Server Installation

1. Install the MCP server:
   ```bash
   go install github.com/mvanhorn/printing-press-library/library/other/hyperliquid/cmd/hyperliquid-pp-mcp@latest
   ```
2. Register with Claude Code:
   ```bash
   claude mcp add hyperliquid-pp-mcp -- hyperliquid-pp-mcp
   ```
3. Verify: `claude mcp list`

## Direct Use

1. Check if installed: `which hyperliquid-pp-cli`
   If not found, offer to install (see Prerequisites at the top of this skill).
2. Match the user query to the best command from the Unique Capabilities and Command Reference above.
3. Execute with the `--agent` flag:
   ```bash
   hyperliquid-pp-cli <command> [subcommand] [args] --agent
   ```
4. If ambiguous, drill into subcommand help: `hyperliquid-pp-cli <command> --help`.
