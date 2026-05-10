# YouTube Script — Trade Hyperliquid by Chatting in Claude Code

Updated for the **Claude Code** demo path (the one we validated end-to-end on
mainnet on May 9). Plain words but slightly less hand-holding than the
Claude Desktop script — Claude Code's audience knows what a terminal is.

Target length: **~3 minutes**.

---

## [0:00 — HOOK — 15 seconds]

🎬 *Camera on you, terminal split-screen with Hyperliquid web app open in a browser tab*

> "I'm going to place a real Hyperliquid leveraged trade by typing one
> sentence to Claude. Watch."

🖥️ *Screen recording: Claude Code chat panel*

> **You type:** *"Use $10 as margin and place a market order 3x long for HYPE."*
> *(Claude calls a few tools, you see "Filled" + position table appear.)*

🎬 *Brief cut to Hyperliquid web UI showing the new position appear*

> "That's a real position. Real money. Real Hyperliquid. From a sentence."

---

## [0:15 — ONE MORE COOL MOMENT — 15 seconds]

🖥️ *Same Claude Code session*

> **You type:** *"What's my PnL on this so far? And what's the funding rate?"*
> *(Claude pulls fills, computes, answers in plain English.)*

> "Read, write, analyze. All from chat. No buttons, no forms."

---

## [0:30 — THE PROMISE — 15 seconds]

🎬 *Camera*

> "Setup is five minutes total. Three things:"
>
> 1. **"Claude Code installed."**
> 2. **"This Hyperliquid CLI from my GitHub."**
> 3. **"A wallet — MetaMask, Rabby, whatever — funded with some USDC on Hyperliquid."**
>
> "Let me show you the whole thing."

---

## [0:45 — STEP 1: INSTALL CLAUDE CODE — 15 seconds]

🖥️ *Terminal*

> "Step one. Install Claude Code if you don't have it."

```bash
npm install -g @anthropic-ai/claude-code
```

> "One command. Then `claude` from anywhere starts a session."

---

## [1:00 — STEP 2: INSTALL THE CLI — 30 seconds]

🖥️ *Terminal*

> "Step two. Get the CLI binary. Pick the file for your platform from the
> latest release."

```bash
# macOS Apple Silicon
curl -L -o /usr/local/bin/hyperliquid \
  https://github.com/ThisNewMark/hyperliquid-pp-cli/releases/latest/download/hyperliquid-darwin-arm64
chmod +x /usr/local/bin/hyperliquid
xattr -d com.apple.quarantine /usr/local/bin/hyperliquid

# Same MCP server binary too
curl -L -o /usr/local/bin/hyperliquid-mcp \
  https://github.com/ThisNewMark/hyperliquid-pp-cli/releases/latest/download/hyperliquid-mcp-darwin-arm64
chmod +x /usr/local/bin/hyperliquid-mcp
xattr -d com.apple.quarantine /usr/local/bin/hyperliquid-mcp
```

> "Linux and Windows binaries are right there too if you're on those."

🖥️ *Verify*

```bash
hyperliquid --version
```

> "If that prints `hyperliquid 1.0.1`, you're good."

---

## [1:30 — STEP 3: BROWSER-BASED SETUP — 60 seconds]

🎬 *Camera*

> "Step three. We're going to make a 'trading agent' — that's a key on your
> computer that can place trades but **cannot** withdraw your money. Then
> we approve it on-chain so Hyperliquid trusts it."

> "Best part: you do the approval through your browser wallet. Your real
> wallet's private key never has to be on this computer. Watch."

🖥️ *Terminal*

```bash
hyperliquid agent setup --user 0xYOUR_MAIN_WALLET_ADDRESS
```

🖥️ *Browser auto-opens to the setup page*

> "Browser opens. The page shows you exactly what's about to be approved —
> the agent address, the fee, what the agent can and can't do."

🖥️ *Highlight the Safety panel briefly*

> "Sign two messages with MetaMask. One for the agent, one for the builder
> fee. That's me — full transparency, one basis point per trade, opt-out
> any time. We'll talk about that in a second."

🖥️ *Two MetaMask popups, click Sign on each*

🖥️ *Terminal shows: "✓ Builder approval detected on-chain. You're done"*

> "Done. The CLI saw the approval land and printed success. Total time:
> about 30 seconds, zero gas fees, no key paste."

---

## [2:30 — STEP 4: WIRE INTO CLAUDE CODE — 15 seconds]

🖥️ *Terminal*

> "Step four. Tell Claude Code about the MCP server."

```bash
claude mcp add hyperliquid /usr/local/bin/hyperliquid-mcp --scope user
```

🖥️ *Verify*

```bash
claude mcp list
```

> "Green check next to `hyperliquid`. Now any Claude Code session can talk to it."

---

## [2:45 — STEP 5: TRADE BY CHATTING — 30 seconds]

🖥️ *Start fresh Claude Code session*

```bash
claude
```

🎬 *Camera, then back to Claude Code*

> "Now you just chat. Try anything."

🖥️ *Run through 4 example chats, fast*

> *"Show me my Hyperliquid positions and PnL."* → Claude pulls them.
>
> *"What's the funding rate on HYPE?"* → Claude pulls it.
>
> *"Use $10 margin and open a 3x long on HYPE, market order."*
> → Claude looks up asset index, current price, sets leverage, places, returns the fill.
>
> *"Close it."* → Claude cancels / closes.

> "Plain English. Claude figures out what you mean and does it."

---

## [3:15 — ABOUT THE FEE — 15 seconds]

🎬 *Camera*

> "Quick honest moment. Each trade through this CLI sends a one basis
> point fee — that's one one-hundredth of a percent — to me. On the
> $30 trade I just did, the fee was about 0.3 cents. If you don't want
> to pay it, just say "place this with no builder fee" — Claude knows
> what to do."

---

## [3:30 — OUTRO — 15 seconds]

🎬 *Camera*

> "Link to download is below. There's also a Claude Desktop bundle for
> drag-drop install if you'd rather chat in a desktop app."

> "Repo, releases, docs — all in the description. Thanks for watching."

🎬 *End card*

---

## Recording notes

- **Set up before recording**: start with a clean state. `agent revoke --force`
  + `agent generate` if you want to demo the setup flow on a fresh agent.
- **Open positions**: close anything you don't want on screen. The cleanest
  recording shows zero positions, then the new trade lands, then close.
- **Don't read the script verbatim**. Hit the beats; improvise.
- **Show real numbers**: real positions, real prices, real builder fees.
  Authenticity is the whole point.
- **Cut aggressively**: any "thinking…" pauses from Claude Code can be
  trimmed in post.

## Alternative title hooks

- "I Trade Crypto by Chatting with Claude. Here's the Setup."
- "Built a Hyperliquid CLI for Claude Code. Trades Lands From One Sentence."
- "Real Hyperliquid Trades from Claude Code in 5 Minutes"

## YouTube description template

> Trade Hyperliquid perpetuals and spot from Claude Code with one sentence
> per trade. Setup takes ~5 minutes. Your main wallet's private key never
> touches your computer — every approval signs in your browser via MetaMask.
>
> Releases (Mac/Linux/Windows binaries): https://github.com/ThisNewMark/hyperliquid-pp-cli/releases
> GitHub: https://github.com/ThisNewMark/hyperliquid-pp-cli
> Setup page: https://thisnewmark.github.io/hyperliquid-pp-cli/setup.html
> Claude Desktop alternative: see QUICKSTART-CLAUDE-DESKTOP.md
>
> ⚠️ Trading crypto is risky. This is a tool for managing your own funds.
> You're responsible for your own decisions. Don't trade what you can't
> afford to lose.
