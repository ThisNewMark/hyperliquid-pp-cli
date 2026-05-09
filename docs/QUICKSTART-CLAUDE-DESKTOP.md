# Quick Start: Claude Desktop

This guide gets you chat-trading on Hyperliquid through Claude Desktop in
under 10 minutes. After setup, you just type things like:

> "What's my BTC position?"
> "Place a $20 long on ETH at the limit price."
> "Cancel any open orders for HYPE."

…and Claude calls the right Hyperliquid tools automatically.

> **Time:** ~10 minutes
> **Cost:** $0 to set up. The two on-chain approvals don't move funds. You'll
> need ~$20 USDC on Hyperliquid for your first demo trade.
> **Private key handling:** your main wallet's private key never touches this
> CLI or this computer's filesystem. All approvals happen via your browser
> wallet (MetaMask, Rabby, etc.).

## Prerequisites

- [Claude Desktop](https://claude.ai/download) installed
- A browser wallet — MetaMask, Rabby, Frame, etc.
- A Hyperliquid account funded with ≥ $20 USDC
  (see [QUICKSTART-TERMINAL.md](QUICKSTART-TERMINAL.md) §2 for the funding flow)

## 1. Install the MCP bundle

The bundle ships as a `.mcpb` file (Model Context Protocol Bundle). Two ways:

**Option A — download from releases** (when published):

Go to the
[printing-press-library releases](https://github.com/mvanhorn/printing-press-library/releases),
download `hyperliquid-mcp-darwin-arm64.mcpb` (or your platform's variant), then
**double-click it**. Claude Desktop will install it automatically.

**Option B — build from source**:

```bash
git clone https://github.com/mvanhorn/printing-press-library.git
cd printing-press-library/library/<path-to-hyperliquid>
make build-mcp
# Then drag bin/hyperliquid-mcp into Claude Desktop's MCP configuration
```

After install, you should see "Hyperliquid" in Claude Desktop's MCP server list
(`Settings → Developer → Edit Config`). **No env vars to configure.**

## 2. Open Claude Desktop and tell it to set up

Restart Claude Desktop after the bundle install. New chat. Tell it:

> **You:** "Set up Hyperliquid for me. My main address is 0xYOUR_MAIN_ADDR."

Claude will call the `agent_setup` MCP tool. It will:

1. Generate a fresh trading agent key on your computer (saved at `~/.hyperliquid/agent.key`, mode 0600 — only you can read it)
2. Print a setup link in the chat
3. Open your browser to a setup page

## 3. Sign two messages in your browser

The setup page shows you exactly what you're about to approve:

- **The agent address** — a fresh key the CLI made; can place trades but cannot withdraw funds
- **The builder address** — the developer's address that gets a tiny per-trade fee
- **The maximum fee rate** — capped at 0.01% (= $0.01 on a $100 trade)
- A **"What this CAN / CANNOT do"** safety panel

Click **Connect Wallet** → MetaMask pops up → approve.

Click **Sign approval #1** → MetaMask pops up → click Sign. The agent is now authorized.

Click **Sign approval #2** → MetaMask pops up → click Sign. The builder fee is approved.

Total: about 30 seconds, two clicks, no key paste, no gas fees.

You'll see "🎉 Setup complete." Close the tab and return to Claude.

## 4. Confirm in Claude

Tell Claude:

> **You:** "Show me builder status for my main wallet."

Claude calls `builder_status` and confirms `approved: true, max_fee_pct:
0.0100%`. Setup done.

## 5. Day-to-day — chat naturally

Try:

> "Show me all my open positions and PnL for the day."
>
> "What's the current funding rate on HYPE?"
>
> "Place a buy on BTC at $50,000, size 0.001, post-only."
>
> "Cancel order 417198341340."
>
> "What was my best trade today?"

Claude routes each through the appropriate MCP tools. Every order carries the
CLI's builder field — you can opt out per-order with "place this trade with no
builder fee."

## What you've got

- **Read access** to your full Hyperliquid account through chat
- **Trade execution** signed by the agent key — Claude can place, modify, and
  cancel orders on your behalf
- **Bounded blast radius** — the agent key on disk can ONLY trade. It cannot
  withdraw, transfer USDC, or approve more agents. Hyperliquid enforces this
  server-side. Your main wallet's private key never touched this computer.

## Troubleshooting

- **Claude says "I don't have access to a Hyperliquid tool"** — fully restart
  Claude Desktop (cmd+Q on macOS) so it picks up the MCP bundle.
- **Browser didn't auto-open** — copy the link Claude printed and paste it
  into your browser manually.
- **MetaMask isn't connecting on the setup page** — make sure MetaMask is
  unlocked and on Arbitrum One (or any chain — the signature works the same
  regardless). If it shows the wrong address, switch accounts in MetaMask.
- **A trade returns "Invalid signature"** — the agent key on disk doesn't
  match what's authorized on-chain. Tell Claude to "regenerate the agent and
  redo setup" — it'll generate a new one and walk you through approval again.
- **"User not registered"** — your main wallet hasn't completed its initial
  USDC deposit. Wait ~2 minutes after depositing, then retry.

## Privacy + safety notes

- The MCP server runs **locally** on your machine. Hyperliquid never sees that
  Claude is the caller — to them it looks like any signed action from your
  agent key.
- Claude Desktop processes your messages; the MCP server only sees structured
  tool calls Claude generates. The bundle and the setup page are open source
  (see the printing-press-library repo).
- Every order carries the CLI's builder field with the developer's address
  authorized at 0.01%. See the README's "Builder Code Transparency" for the
  full disclosure.
- The agent key on your laptop is stored at `~/.hyperliquid/agent.key` mode
  0600 (only your user can read it). To rotate, tell Claude "rotate my agent
  key" or run `hyperliquid agent revoke --force` followed by `hyperliquid
  agent setup` in a terminal.

## See also

- [QUICKSTART-TERMINAL.md](QUICKSTART-TERMINAL.md) — same setup, different
  surface (terminal-only, no Claude Desktop)
- [README.md](../README.md) — full command reference and Builder Code
  Transparency disclosure
