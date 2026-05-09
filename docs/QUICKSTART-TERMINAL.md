# Quick Start: Terminal & Claude Code

Zero to placing a real Hyperliquid order through the `hyperliquid` CLI in
under 10 minutes. Works in any terminal (Terminal.app, iTerm, Warp, the
embedded terminal in Claude Code, etc.).

> **Time:** ~10 minutes
> **Cost:** $0 to set up. Two on-chain approvals don't move funds. You'll
> need ~$20 USDC on Hyperliquid for the demo trade.
> **Private key handling:** your main wallet's private key never gets typed
> into a terminal or saved to disk. All approvals happen via your browser
> wallet.

## Prerequisites

- macOS, Linux, or Windows (WSL works fine)
- A browser wallet (MetaMask, Rabby, Frame) with some USDC on Arbitrum One
- Either Go 1.26+ (for build-from-source) or a pre-built binary

## 1. Install the CLI

### Option A — pre-built binary (fastest)

Download for your platform from the
[GitHub releases](https://github.com/ThisNewMark/hyperliquid-pp-cli/releases):

```bash
# macOS Apple Silicon
curl -L -o /usr/local/bin/hyperliquid \
  https://github.com/ThisNewMark/hyperliquid-pp-cli/releases/latest/download/hyperliquid-darwin-arm64
chmod +x /usr/local/bin/hyperliquid
xattr -d com.apple.quarantine /usr/local/bin/hyperliquid  # macOS Gatekeeper

# macOS Intel
curl -L -o /usr/local/bin/hyperliquid \
  https://github.com/ThisNewMark/hyperliquid-pp-cli/releases/latest/download/hyperliquid-darwin-amd64
chmod +x /usr/local/bin/hyperliquid

# Linux
curl -L -o /usr/local/bin/hyperliquid \
  https://github.com/ThisNewMark/hyperliquid-pp-cli/releases/latest/download/hyperliquid-linux-amd64
chmod +x /usr/local/bin/hyperliquid

# Windows: download hyperliquid-windows-amd64.exe and add to PATH
```

### Option B — build from source (Go 1.26.3+)

```bash
git clone https://github.com/ThisNewMark/hyperliquid-pp-cli.git
cd hyperliquid-pp-cli
make build
sudo cp bin/hyperliquid /usr/local/bin/
```

### Verify

```bash
hyperliquid --version
# Expect: hyperliquid 1.0.0

hyperliquid doctor
# Expect: Config OK, API reachable
```

## 2. Fund a Hyperliquid wallet (skip if you already have an account)

If you already have a Hyperliquid account funded with ≥ $20 USDC, skip ahead.

Otherwise:

1. Open MetaMask → create a new account for this demo (recommended for
   bounded risk). Note the 0x address.
2. Send **~$20 USDC on Arbitrum One** to that address (Hyperliquid bridges
   from Arbitrum, not Ethereum mainnet).
3. Go to https://app.hyperliquid.xyz, click **Connect** with the new
   account, then **Deposit**. Approve in MetaMask. Takes ~1 minute.

## 3. Run the one-shot setup command

```bash
hyperliquid agent setup --user 0xYOUR_MAIN_ADDRESS
```

This will:

1. Generate a fresh trading agent key on your computer (saved at
   `~/.hyperliquid/agent.key` mode 0600). The agent can trade but cannot
   withdraw or transfer funds.
2. Open your browser to a setup page that shows you exactly what's about
   to be approved.
3. Wait for you to sign two MetaMask popups — one to authorize the agent,
   one to approve the builder fee.
4. Poll Hyperliquid until the approvals land, then print success.

The `--user` flag tells the CLI which Hyperliquid account to watch for the
approval. **It's just your public address — no private info.**

### What the browser page does

When the page opens, you'll see:

- The agent address (matches what your CLI just printed — verify they match)
- The builder address that will receive fees
- The maximum fee rate (0.01% by default — 1 basis point)
- A "What this CAN / CANNOT do" safety panel

Click **Connect Wallet** → MetaMask pops up → approve.

Click **Sign approval #1** → MetaMask pops up showing the typed-data message
→ click Sign. The agent is now authorized.

Click **Sign approval #2** → MetaMask pops up → click Sign. The builder fee
is approved.

Total: ~30 seconds, two clicks, no gas fees, no private key paste.

When you see "🎉 Setup complete." in the browser, return to your terminal.
The CLI's polling will pick up the approval within a few seconds and print:

```
✓ Builder approval detected on-chain. You're done — start trading with `hyperliquid exchange place-order`.
```

## 4. Place your first order

The agent key handles trading from now on. No more env vars needed.

```bash
# Tiny ALO (post-only) buy 50% below mark — sits on the book, can't fill
hyperliquid exchange place-order \
  --orders '[{"a":0,"b":true,"p":"30000","s":"0.001","r":false,"t":{"limit":{"tif":"Alo"}}}]'
```

Expected response (the `oid` is the order ID):

```json
{
  "data": {
    "response": {
      "data": { "statuses": [ { "resting": { "oid": 417198341340 } } ] },
      "type": "order"
    },
    "status": "ok"
  },
  "signer": "0xa063...f6c3",
  "status": 200,
  "success": true
}
```

Check it on https://app.hyperliquid.xyz — the order appears under your open
orders.

## 5. Cancel it

```bash
hyperliquid exchange cancel-orders --cancels '[{"a":0,"o":417198341340}]'
```

(replace `417198341340` with your `oid`).

## Per-asset precision (important for trading)

Each Hyperliquid asset has a `szDecimals` value (returned by
`info get-perp-meta` / `info get-spot-meta`) that limits how many decimal
places your price and size can have. Send too many and Hyperliquid rejects
with `"Price must be divisible by tick size"` or
`"Size has too many decimal places"`.

The rules:

| Field | Max decimal places |
|---|---|
| **Size** | `szDecimals` |
| **Price (perps)** | `6 - szDecimals` |
| **Price (spot)** | `8 - szDecimals` |

Plus a hard cap of **5 significant figures** on prices.

Examples:

| Asset | szDecimals | Valid size | Valid price |
|---|---|---|---|
| BTC perp | 5 | `0.00012` | `70250.5` |
| ETH perp | 4 | `0.0123` | `3450.25` |
| HYPE perp | 2 | `0.57` | `42.9525` |

The CLI's signing pipeline auto-strips trailing zeros (so `"5.0"` becomes
`"5"` on the wire — required for canonical msgpack), but it does **not**
auto-round to the asset's tick. If you over-specify decimals, you'll get a
clean Hyperliquid error and can retry with the right precision.

Quick way to check an asset's szDecimals:

```bash
hyperliquid info get-perp-meta --json | jq '.data.universe[] | select(.name=="HYPE") | {name, szDecimals}'
```

## What's next

You're set up:

- An agent wallet that can trade indefinitely without your main key on disk
- An on-chain approval crediting the CLI's builder address on every order
- The full read + write surface — try `hyperliquid info get-clearinghouse-state --user 0xYOUR_MAIN_ADDRESS` for positions, or `hyperliquid info get-all-mids` for live prices

For programmatic use: every command supports `--json` for clean output and
`--agent` for non-interactive defaults (no prompts, JSON, no color). Both
make the CLI easy to wire into shell scripts, cron jobs, or Python.

For LLM use: every visible command is also exposed as an MCP tool — see
[QUICKSTART-CLAUDE-DESKTOP.md](QUICKSTART-CLAUDE-DESKTOP.md) for chat-trading
through Claude Desktop.

## Manual flow (no browser)

If your environment can't open a browser (e.g., a remote server), you can do
the setup the old way using a private key. Not recommended for production
machines:

```bash
export HYPERLIQUID_MAIN_KEY="0xYOUR_PRIVATE_KEY"   # one-time
hyperliquid agent generate
hyperliquid agent approve
hyperliquid builder approve
unset HYPERLIQUID_MAIN_KEY
```

Same end result, but the main key briefly touches your shell environment.
The browser flow is preferred whenever possible.

## Troubleshooting

- **Browser didn't auto-open** — the `setup` command prints the URL; copy it
  to your browser manually. Use `--no-open` to suppress the auto-open
  attempt.
- **`agent setup` says "User not registered"** — your main wallet hasn't
  finished its USDC deposit yet. Wait ~2 minutes and rerun.
- **MetaMask shows wrong account** — switch accounts in the extension before
  clicking "Connect Wallet" on the setup page.
- **`place-order` returns "minimum value of $10"** — Hyperliquid requires
  ≥$10 notional. With BTC at ~$70K, `s=0.0002` (~$14) works.
- **"Invalid signature" on a real order** — the agent key on disk doesn't
  match what's authorized on-chain. Run `hyperliquid agent revoke --force`
  followed by `hyperliquid agent setup --user 0xYOUR_ADDR` to start over.
