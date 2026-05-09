# Hyperliquid CLI

Hyperliquid is an on-chain order-book L1 perpetuals + spot exchange. The
public HTTP API exposes two physical endpoints — `/info` (read) and
`/exchange` (write) — that multiplex many logical actions via a `type`
discriminator in the request body.

For CLI ergonomics, this spec models each discriminator value as its own
path. The generated client wraps each call in the appropriate envelope
targeting `/info` or `/exchange` against the configured network base URL.

Learn more at [Hyperliquid](https://hyperliquid.gitbook.io/hyperliquid-docs).

## Builder Code Transparency

This CLI ships with a default **builder code** — the address that receives a
small per-trade fee on orders placed through it. Hyperliquid's builder-code
mechanism is consent-based: every fee-bearing order requires you to have
on-chain-approved the builder address with a maximum fee rate. Until you do
that approval, no fee is charged regardless of what flags you pass.

### What you'll see on the wire

Every order this CLI submits via `exchange place-order` carries an extra field:

```json
"builder": { "b": "<builder address>", "f": <fee in tenths of basis points> }
```

The fee unit is **tenths of basis points**: `f=10` means 1bp = 0.01%. Server
caps the fee at 0.1% on perps and 1% on spot.

### The shipping defaults

| Setting | Value | Source |
|---|---|---|
| Default builder address | `0xc8f0cd137e28f717a20f810b46926f92978bbcfa` | `internal/builder/builder.go::DefaultBuilderAddress` |
| Default fee | `10` (tenths of bps = 0.01%, i.e. 1 basis point) | `internal/builder/builder.go::DefaultBuilderFeeBps` |
| Default `maxFeeRate` (for `builder approve`) | `0.01%` | `internal/builder/builder.go::DefaultMaxFeeRate` |

The default fee is **0.01%** — 1 basis point, the lowest end of trading fees
and well below Hyperliquid's 0.1% perps cap. Opt out of builder fees entirely
at any time with `--no-builder` or `--builder 0x0`.

### How to opt out of builder fees

Three equivalent ways:

```bash
# Per-order opt-out (recommended for transparency in scripts)
hyperliquid exchange place-order --no-builder ...

# Equivalent: explicit zero address
hyperliquid exchange place-order --builder 0x0 ...

# Or set your own address
hyperliquid exchange place-order --builder 0xYourAddress --builder-fee-bps 10 ...
```

### How to verify the builder address on-chain

You can independently confirm what address this CLI is configured to credit:

```bash
hyperliquid builder status --user <your-address>
```

That command queries Hyperliquid's `info {type:approvedBuilders}` and
`info {type:maxBuilderFee}` endpoints and shows whether you have approved this
CLI's default builder, and at what cap.

You can also verify directly without trusting the CLI by curling the public API:

```bash
curl -s -X POST https://api.hyperliquid.xyz/info \
  -H 'content-type: application/json' \
  -d '{"type":"approvedBuilders","user":"<your-address>"}'
```

### How to approve

```bash
hyperliquid builder approve --builder <addr> --max-fee-rate 0.01%
```

This action **must be signed by your main depositing wallet, not an agent
wallet**. Hyperliquid enforces this server-side. To revoke later, use
`hyperliquid builder revoke --builder <addr>`, which sends an
`approveBuilderFee` action with `maxFeeRate=0%`.

## Install

### Pre-built binary (fastest)

Download for your platform from the [latest GitHub release](https://github.com/ThisNewMark/hyperliquid-pp-cli/releases) and put it on your `PATH`:

```bash
# macOS Apple Silicon
curl -L -o /usr/local/bin/hyperliquid \
  https://github.com/ThisNewMark/hyperliquid-pp-cli/releases/latest/download/hyperliquid-darwin-arm64
chmod +x /usr/local/bin/hyperliquid
xattr -d com.apple.quarantine /usr/local/bin/hyperliquid

# macOS Intel
curl -L -o /usr/local/bin/hyperliquid \
  https://github.com/ThisNewMark/hyperliquid-pp-cli/releases/latest/download/hyperliquid-darwin-amd64
chmod +x /usr/local/bin/hyperliquid

# Linux (amd64)
curl -L -o /usr/local/bin/hyperliquid \
  https://github.com/ThisNewMark/hyperliquid-pp-cli/releases/latest/download/hyperliquid-linux-amd64
chmod +x /usr/local/bin/hyperliquid

# Windows: download hyperliquid-windows-amd64.exe and add to PATH
```

### Build from source (Go 1.26.3+)

```bash
git clone https://github.com/ThisNewMark/hyperliquid-pp-cli.git
cd hyperliquid-pp-cli
make build
sudo cp bin/hyperliquid /usr/local/bin/
```

### Build all platforms (for releases)

```bash
make release      # produces dist/hyperliquid-{darwin-arm64,darwin-amd64,linux-amd64,linux-arm64,windows-amd64.exe}
make release-mcp  # MCP server binaries for the same platforms
```

### Verify

```bash
hyperliquid --version    # expect: hyperliquid 1.0.0
hyperliquid doctor       # expect: Config OK, API reachable
```

## Quick Start

Two paths from zero to a real trade in 10 minutes — both walked through end-to-end:

- **[Terminal / Claude Code](docs/QUICKSTART-TERMINAL.md)** — for command-line users
- **[Claude Desktop](docs/QUICKSTART-CLAUDE-DESKTOP.md)** — for chat-trading via MCP

The setup flow uses a browser-based signing page. **Your wallet's private key never gets typed into a terminal or saved to disk** — every approval happens via your browser wallet (MetaMask, Rabby, Frame, etc.).

## Usage

Run `hyperliquid --help` for the full command reference and flag list.

## Commands

### exchange

Manage exchange

- **`hyperliquid exchange approve-agent`** - Approve an agent (API) wallet for trading
- **`hyperliquid exchange approve-builder-fee`** - MUST be signed by the user's main wallet, not an agent. Max fee rate is
a human-readable percentage string like "0.001%". Server caps: 0.1% on
perps, 1% on spot. Each user may have up to 10 active approvals.
- **`hyperliquid exchange batch-modify-orders`** - Modify multiple orders atomically
- **`hyperliquid exchange cancel-orders`** - Cancel orders by oid
- **`hyperliquid exchange cancel-orders-by-cloid`** - Cancel orders by client order id
- **`hyperliquid exchange cancel-twap-order`** - Cancel a TWAP order
- **`hyperliquid exchange class-transfer`** - Move USDC between perp and spot
- **`hyperliquid exchange delegate-token`** - Delegate (or undelegate) staked HYPE to a validator
- **`hyperliquid exchange modify-order`** - Modify a single order
- **`hyperliquid exchange place-order`** - Place one or more orders. Optionally attaches a builder code.
- **`hyperliquid exchange place-twap-order`** - Place a TWAP order
- **`hyperliquid exchange schedule-cancel`** - Dead-man-switch — cancel all orders at a future timestamp
- **`hyperliquid exchange send-spot`** - Transfer a spot asset
- **`hyperliquid exchange send-usd`** - Transfer USDC on the perps account
- **`hyperliquid exchange stake-deposit`** - Stake HYPE into the staking layer
- **`hyperliquid exchange stake-withdraw`** - Begin a 7-day unstake from the staking layer
- **`hyperliquid exchange update-isolated-margin`** - Add/remove isolated margin on an asset
- **`hyperliquid exchange update-leverage`** - Update cross/isolated leverage for an asset
- **`hyperliquid exchange vault-transfer`** - Deposit to or withdraw from a vault
- **`hyperliquid exchange withdraw`** - Withdraw USDC to L1 (~5min, $1 fee)

### info

Read-only queries against /info

- **`hyperliquid info get-all-mids`** - All mid prices
- **`hyperliquid info get-approved-builders`** - List of builder addresses this user has approved
- **`hyperliquid info get-candle-snapshot`** - Candles for a coin and interval
- **`hyperliquid info get-clearinghouse-state`** - Perp positions, margin summary, withdrawable
- **`hyperliquid info get-frontend-open-orders`** - User's open orders (frontend form, includes triggers)
- **`hyperliquid info get-funding-history`** - Historical funding rates for a coin
- **`hyperliquid info get-historical-orders`** - Recent historical orders
- **`hyperliquid info get-l2-book`** - L2 order book for a coin
- **`hyperliquid info get-max-builder-fee`** - Max builder fee in tenths of bps that this user has approved for this builder
- **`hyperliquid info get-open-orders`** - User's open orders (compact form)
- **`hyperliquid info get-order-status`** - Status for a single order by oid or cloid
- **`hyperliquid info get-perp-meta`** - Get perpetuals universe and margin tables
- **`hyperliquid info get-perp-meta-and-asset-ctxs`** - Perp metadata plus per-asset mark/funding context
- **`hyperliquid info get-perps-at-open-interest-cap`** - Coins currently at OI cap
- **`hyperliquid info get-portfolio`** - Account value and PnL history (day/week/month/allTime)
- **`hyperliquid info get-predicted-fundings`** - Predicted next funding rates across venues
- **`hyperliquid info get-referral`** - Referral state including builder rewards
- **`hyperliquid info get-spot-clearinghouse-state`** - Spot balances
- **`hyperliquid info get-spot-meta`** - Spot universe metadata
- **`hyperliquid info get-spot-meta-and-asset-ctxs`** - Spot metadata plus per-asset context
- **`hyperliquid info get-sub-accounts`** - Sub-accounts owned by this address
- **`hyperliquid info get-user-fees`** - User fee schedule and current rates
- **`hyperliquid info get-user-fills`** - User fills (most recent ~2000)
- **`hyperliquid info get-user-fills-by-time`** - User fills in a time range
- **`hyperliquid info get-user-funding`** - User funding payment history
- **`hyperliquid info get-user-non-funding-ledger-updates`** - Deposits, withdrawals, transfers (non-funding ledger)
- **`hyperliquid info get-user-role`** - Identify whether an address is user/agent/vault/subAccount/missing
- **`hyperliquid info get-user-vault-equities`** - User's equity in each vault
- **`hyperliquid info get-vault-details`** - Details for a single vault


## Output Formats

```bash
# Human-readable table (default in terminal, JSON when piped)
hyperliquid exchange approve-agent --agent-address example-value

# JSON for scripting and agents
hyperliquid exchange approve-agent --agent-address example-value --json

# Filter to specific fields
hyperliquid exchange approve-agent --agent-address example-value --json --select id,name,status

# Dry run — show the request without sending
hyperliquid exchange approve-agent --agent-address example-value --dry-run

# Agent mode — JSON + compact + no prompts in one flag
hyperliquid exchange approve-agent --agent-address example-value --agent
```

## Agent Usage

This CLI is designed for AI agent consumption:

- **Non-interactive** - never prompts, every input is a flag
- **Pipeable** - `--json` output to stdout, errors to stderr
- **Filterable** - `--select id,name` returns only fields you need
- **Previewable** - `--dry-run` shows the request without sending
- **Explicit retries** - add `--idempotent` to create retries when a no-op success is acceptable
- **Confirmable** - `--yes` for explicit confirmation of destructive actions
- **Piped input** - write commands can accept structured input when their help lists `--stdin`
- **Offline-friendly** - sync/search commands can use the local SQLite store when available
- **Agent-safe by default** - no colors or formatting unless `--human-friendly` is set

Exit codes: `0` success, `2` usage error, `3` not found, `5` API error, `7` rate limited, `10` config error.

## Use with Claude Code

Install the focused skill — it auto-installs the CLI on first invocation:

```bash
npx skills add mvanhorn/printing-press-library/cli-skills/pp-hyperliquid -g
```

Then invoke `/pp-hyperliquid <query>` in Claude Code. The skill is the most efficient path — Claude Code drives the CLI directly without an MCP server in the middle.

<details>
<summary>Use as an MCP server in Claude Code (advanced)</summary>

If you'd rather register this CLI as an MCP server in Claude Code, install the MCP binary first:

```bash
go install github.com/mvanhorn/printing-press-library/library/other/hyperliquid/cmd/hyperliquid-pp-mcp@latest
```

Then register it:

```bash
claude mcp add hyperliquid hyperliquid-pp-mcp
```

</details>

## Use with Claude Desktop

This CLI ships an [MCPB](https://github.com/modelcontextprotocol/mcpb) bundle — Claude Desktop's standard format for one-click MCP extension installs (no JSON config required).

To install:

1. Download the `.mcpb` for your platform from the [latest release](https://github.com/mvanhorn/printing-press-library/releases/tag/hyperliquid-current).
2. Double-click the `.mcpb` file. Claude Desktop opens and walks you through the install.

Requires Claude Desktop 1.0.0 or later. Pre-built bundles ship for macOS Apple Silicon (`darwin-arm64`) and Windows (`amd64`, `arm64`); for other platforms, use the manual config below.

<details>
<summary>Manual JSON config (advanced)</summary>

If you can't use the MCPB bundle (older Claude Desktop, unsupported platform), install the MCP binary and configure it manually.

```bash
go install github.com/mvanhorn/printing-press-library/library/other/hyperliquid/cmd/hyperliquid-pp-mcp@latest
```

Add to your Claude Desktop config (`~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "hyperliquid": {
      "command": "hyperliquid-pp-mcp"
    }
  }
}
```

</details>

## Health Check

```bash
hyperliquid doctor
```

Verifies configuration and connectivity to the API.

## Configuration

Config file: `~/.config/hyperliquid/config.toml`

## Troubleshooting
**Not found errors (exit code 3)**
- Check the resource ID is correct
- Run the `list` command to see available items

## HTTP Transport

This CLI uses Chrome-compatible HTTP transport for browser-facing endpoints. It does not require a resident browser process for normal API calls.

---

Generated by [CLI Printing Press](https://github.com/mvanhorn/cli-printing-press)
