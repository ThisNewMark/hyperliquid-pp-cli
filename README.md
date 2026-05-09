# Hyperliquid CLI

A Go CLI and MCP server for [Hyperliquid](https://hyperliquid.gitbook.io/hyperliquid-docs)
perpetuals and spot trading. Trade from your terminal or chat with Claude
Desktop. Setup uses a browser-based signing flow — your main wallet's private
key never touches this computer.

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

## Use with Claude Desktop

The CLI ships an [MCPB](https://github.com/modelcontextprotocol/mcpb) bundle —
Claude Desktop's standard format for one-click MCP extension installs.

1. Download the `.mcpb` for your platform from the [latest release](https://github.com/ThisNewMark/hyperliquid-pp-cli/releases).
2. Double-click the `.mcpb` file. Claude Desktop walks you through the install.
3. Run `/agent setup` from a chat or follow [QUICKSTART-CLAUDE-DESKTOP.md](docs/QUICKSTART-CLAUDE-DESKTOP.md).

Requires Claude Desktop 1.0.0 or later. Pre-built bundles ship for macOS
Apple Silicon, macOS Intel, Linux (amd64/arm64), and Windows.

<details>
<summary>Manual MCP config (advanced)</summary>

```bash
go install github.com/ThisNewMark/hyperliquid-pp-cli/cmd/hyperliquid-pp-mcp@latest
```

Add to `~/Library/Application Support/Claude/claude_desktop_config.json`:

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

## Builder fee

Orders placed through this CLI default to a 1 basis point (0.01%) builder fee
that goes to the developer's address (`0xc8f0cd137e28f717a20f810b46926f92978bbcfa`).
Opt out per-order with `--no-builder` or `--builder 0x0`. Verify what's
authorized with `hyperliquid builder status --user 0xYOUR_ADDR`. Full details
on mechanics, opt-out, and on-chain verification in
[docs/BUILDER-FEE.md](docs/BUILDER-FEE.md).

## Troubleshooting

**Not found errors (exit code 3)**
- Check the resource ID is correct
- Run the relevant `info get-*` command to see available items

---

Generated by [CLI Printing Press](https://github.com/mvanhorn/cli-printing-press)
