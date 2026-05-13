# Hyperliquid CLI — Video & Project Ideas

Companion ideas backlog for things to build / record / promote *beyond* the
core "chat-trade with Claude" demo. Pick from this list when ready for the
next move.

---

## 🎬 Video ideas (post-launch)

### 1. The launch demo (recording now)
**File:** [YOUTUBE-SCRIPT-CLAUDE-CODE.md](YOUTUBE-SCRIPT-CLAUDE-CODE.md)
**Length:** ~3 min
**Status:** primary release video — chat-trades a real Hyperliquid position via Claude Code.

### 2. "Trade Hyperliquid through Claude Desktop" (alt path)
**File:** [YOUTUBE-SCRIPT-CLAUDE-DESKTOP.md](YOUTUBE-SCRIPT-CLAUDE-DESKTOP.md)
**Length:** ~3 min
**Why second:** Claude Desktop has Anthropic's policy filter. Worth re-testing now that signing is fixed; if it works, this is a more mainstream-accessible demo (drag-drop install, no terminal). If still blocked, skip this one.

### 3. "Build a funding-rate arb bot in 30 min" (Project #4)
**Length:** 5–10 min
**Audience:** quant-curious devs
**Hook:** *"AI agents will arb funding rates better than you. Here's a working one in 50 lines."*
**Drives:** demonstrates the CLI + MCP in a non-trivial autonomous loop

### 4. "Hedge any token with a perp short via chat" (Project #5)
**Length:** 3–5 min
**Audience:** Hyperliquid spot traders, HIP-1 token holders
**Hook:** *"You hold $5K of [token]. The market dumps tomorrow. Here's how Claude hedges in 1 sentence."*
**Drives:** captures the niche-but-passionate token-holder segment

### 5. "Trading as a service for your friends" (Project #7)
**Length:** 5–8 min
**Audience:** community/Discord operators
**Hook:** *"Run a chat-trade box for your group. Each friend's agent wallet, your builder fee."*
**Drives:** direct revenue multiplier — every viewer who sets this up becomes a multi-user builder fee channel

### 6. "Why your AI agent can't drain your account" (safety explainer)
**Length:** 2–3 min
**Audience:** crypto-cautious / security-minded
**Hook:** Walk through the agent-wallet model — what an attacker could and couldn't do if your laptop got hacked.
**Drives:** removes the #1 objection ("AI + my wallet = scary")

### 7. "Day in the life — chat-trading Hyperliquid"
**Length:** 5–10 min
**Audience:** trader/lifestyle vlog crowd
**Format:** record a real trading day where every interaction is via Claude Code. PnL at the end.
**Drives:** social proof; very organic for X/Twitter clips

---

## 🛠️ Project ideas (things to build on top)

### 1. Voice-controlled trading (consumer, magical) ⭐
- **Pitch:** Hold a key, say *"buy $100 of SOL perp"*, Claude executes.
- **Stack:** Whisper or macOS dictation → Claude Code (or direct Anthropic SDK) → our MCP.
- **Audience:** anyone who liked the demo but doesn't want to type.
- **Effort:** ~1 day; raycast/shortcut extension is the polish layer.

### 2. Daily / weekly portfolio email reports (light user)
- **Pitch:** Cron + Claude → 5-line PnL summary in your inbox.
- **Stack:** cron job → `claude` invocation with a system prompt → SMTP / SendGrid → email.
- **Audience:** casual traders who want passive insight.
- **Effort:** half-day.

### 3. Twitter/Discord signal → confirm-then-execute
- **Pitch:** Bot watches a channel; when signal matches your rules, Slacks you a confirm button; click yes → Claude trades.
- **Stack:** Discord/Twitter API → Claude → MCP → Slack webhook for confirmation.
- **Audience:** signal followers who want a human-in-the-loop.
- **Effort:** ~2 days.

### 4. Funding-rate arbitrage bot (sophisticated) ⭐
- **Pitch:** Hourly: identify highest-paying-funding shorts, open hedged positions, harvest funding.
- **Stack:** scheduler → Claude with strategy prompt → MCP for execution → local SQLite for state.
- **Audience:** quant-leaning traders.
- **Effort:** ~3 days; pair with PerpLobster for the bot framework.
- **Cross-promo:** ties into PerpLobster's existing audience.

### 5. Hedge any token with a perp short (token holder) ⭐
- **Pitch:** Claude watches your spot balance; auto-opens a perp short of the same notional → delta-neutral with funding income.
- **Stack:** wallet-state poller → Claude → MCP for hedge sizing.
- **Audience:** HIP-1 token holders, memecoin bagholders.
- **Effort:** ~2 days. Could be the same friend group from #7.

### 6. AI co-pilot inside an existing dashboard (web dev)
- **Pitch:** TradingView-style chart UI + a chat panel that runs Claude with our MCP. User clicks a chart, asks *"long this level with a 5% stop"*, executes.
- **Stack:** Next.js / SvelteKit + Claude API + MCP server on a backend.
- **Audience:** trading-tool-startup founders / web devs.
- **Effort:** ~1 week (real product).

### 7. Multi-user trading-as-a-service for friends/family ⭐⭐ (highest-leverage)
- **Pitch:** Discord bot. Each user has their own agent wallet. They chat-trade their own funds. You collect builder fees on every trade.
- **Stack:** Discord bot wrapping Claude → per-user MCP sessions / agent-key files → our CLI.
- **Audience:** Discord/Telegram community operators.
- **Effort:** ~3 days for v0; multi-user infra is the bulk.
- **Why this:** direct revenue multiplier per user added.

### 8. Strategy backtester + live executor (quant)
- **Pitch:** Local Python backtesting on Hyperliquid candles + live execution via our CLI when the strategy fires.
- **Stack:** Python with hyperliquid SDK for backtest data, our CLI as a subprocess for live trades.
- **Audience:** quant traders, crypto-strategy developers.
- **Effort:** ~1 week if you want a real workflow.

### 9. Liquidation watcher with auto-action
- **Pitch:** Polls `clearinghouseState`; when liquidation distance drops below threshold, Claude messages you OR auto-deleverages OR adds margin from spot.
- **Stack:** background process → Claude with risk prompt → MCP execution.
- **Audience:** leveraged traders who don't want to babysit.
- **Effort:** ~2 days.

### 10. Crypto research agent that *also* trades (newsletter)
- **Pitch:** Daily research report ("why HYPE looks overextended") with an executable trade idea. Subscribers chat-execute via their own setup.
- **Stack:** RSS / Twitter / on-chain data → Claude → markdown report → embed CLI command.
- **Audience:** newsletter writers, alpha generators.
- **Effort:** ~1 week (research framework is the work).

### 11. HIP-3 support — builder-deployed perps (tokenized stocks, commodities) ⭐
- **Pitch:** Same CLI/MCP, expanded asset universe. *"Long $50 of tokenized Tesla at 2x via chat."*
- **Why:** HIP-3 markets have **$1.43B OI as of May 2026** — 23 of Hyperliquid's top 30 trading pairs are HIP-3 (tokenized stocks, commodities, indices). Our CLI currently only supports standard perps; missing this is missing the more interesting half of the platform.
- **What's missing:** `dex` field in `PlaceOrderAction` and `OrderWire`; `--dex` CLI flag; surfacing the `perpDexs` / `perpDexLimits` info endpoints; MCP context tool note about asset-naming convention (`xyz:GOLD`, `flx:XMR`).
- **Effort:** ~1 day. Mostly mechanical — same signing path, just a new field.
- **Doc:** https://hyperliquid.gitbook.io/hyperliquid-docs/hyperliquid-improvement-proposals-hips/hip-3-builder-deployed-perpetuals
- **Video angle:** *"You don't have to trade BTC. Tokenized stocks. Commodities. Same chat. Same flow."*

### 12. HIP-4 support — outcome markets / prediction markets ⭐⭐ (first-mover potential)
- **Pitch:** Chat-bet on real-world outcomes via Hyperliquid's HIP-4 surface. YES/NO tokens, fully collateralized, no liquidation.
- **Why:** HIP-4 launched **May 2, 2026** — only days old when we're writing this. Hyperliquid is going after Polymarket. Building the first AI chat client for HIP-4 means owning the "I told Claude to bet on the election" demo for the next year.
- **What's missing:** Entirely new action types (not yet read the spec). New info endpoints. New asset model (YES/NO tokens vs perp pairs). New CLI tree — probably `hyperliquid outcome ...`.
- **Effort:** ~2-3 days. The spec read is the longest part.
- **Doc:** https://hyperliquid.gitbook.io/hyperliquid-docs/hyperliquid-improvement-proposals-hips/hip-4-outcome-markets
- **Video angle:** *"I bet on the next CPI print by chatting with Claude. Here's how."*

---

## ⚡ Weekend hacks (light projects, big wins)

- **Slack/Discord bot read-only** — `/hl pnl` returns daily PnL. No key handling. ~3 hours.
- **iOS Shortcut** — homescreen widget that pings a Mac running Claude → reads positions out loud. ~2 hours.
- **Custom MCP descriptions fork** — *"degenmode"* version with rougher tool descriptions and bigger default sizes. ~1 hour.
- **OpenClaw skill wrapper** — `/hyperliquid` skill on ClawHub for one-command install. Mirror PerpLobster's pattern. ~half-day.
- **`hyperliquid status` aggregated view** — single command that prints positions + open orders + spot balance + funding accrued + builder rewards in one nice ASCII table. ~3 hours.
- **Automated builder-fee revenue dashboard** — pull `info_get-referral` daily, log builderRewards over time, simple Grafana / line chart. ~half-day.

---

## 🎯 First moves (priority order)

When you come back to this list:

1. **Record + ship the launch video** (you're doing this now)
2. **Test in Claude Desktop again** now that signing works — if it goes through, record video #2 to capture that audience
3. **Build #7 (multi-user TaaS)** — direct revenue multiplier, ~3 days
4. **Record video #3 demoing the bot from #4 (funding arb)** — drives the quant audience
5. **Add OpenClaw skill wrapper (weekend hack)** — gets you in front of the ClawHub audience that already trusts your PerpLobster work
6. **Build #5 (token holder hedger)** — captures HIP-1 token communities

Don't try to do everything. Pick #2 or #3 next based on whether you have time for video work or coding work this week.

---

## 🏷️ Compatibility notes (works with these MCP-aware tools)

The MCP server in this repo (`bin/hyperliquid-mcp`) speaks the standard
Model Context Protocol. Any MCP-compliant client can use it. Confirmed
or expected:

| Tool | Setup pattern |
|---|---|
| **Claude Code** | `claude mcp add hyperliquid /path/to/hyperliquid-mcp --scope user` ✅ validated mainnet end-to-end 2026-05-09 |
| **Claude Desktop** | Drag-drop the `.mcpb` file (auto-installs) — re-test post-signing-fix |
| **Hermes** (Anthropic) | MCP config same shape as Claude Code |
| **OpenClaw / ClawHub** | MCP entry in their config; alternatively wrap as a `/hyperliquid` skill |
| **Cursor IDE** | Edit `~/.cursor/mcp.json` to point at `hyperliquid-mcp` |
| **Cline** (VS Code) | Add to Cline's MCP settings |
| **Codex CLI** (OpenAI) | `codex mcp add` or `~/.codex/config.toml` |
| **Continue.dev** | MCP entry in `~/.continue/config.yaml` |
| **Goose** (Block) | MCP entry in their config |
| **ChatGPT** (consumer) | ⚠️ Limited MCP support; rolling out |
| **Perplexity Comet** | ⚠️ Their guidance discourages MCP for write actions |

Same binary, different config files. The `.mcpb` bundle is Claude
Desktop's specific drag-drop format; everywhere else uses the raw
`hyperliquid-mcp` binary directly.
