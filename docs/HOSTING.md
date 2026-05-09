# Hosting the setup page

`docs/setup.html` is a single static HTML page that handles the one-time
on-chain authorizations for the Hyperliquid CLI:

1. `approveAgent` — registers the CLI's agent wallet for trading
2. `approveBuilderFee` — registers the CLI's builder code at a max fee rate

Both actions are signed via the user's browser wallet (MetaMask, Rabby, etc.)
using EIP-712 typed-data signatures. **The user's private key never leaves
their wallet.**

## How the CLI uses it

The CLI's `agent setup` command:

1. Generates an agent key locally (if missing)
2. Constructs a URL with the agent address, builder address, and max fee rate
   as query parameters
3. Auto-opens that URL in the user's default browser
4. Polls Hyperliquid's `/info {type:approvedBuilders}` until the approval
   lands, then prints success

## URL format

```
https://<your-host>/setup.html?agent=0x<40hex>&builder=0x<40hex>&maxFee=0.01%25&network=mainnet
```

| Param | Required | Description |
|-------|----------|-------------|
| `agent` | Yes | Agent wallet address to authorize (lowercase 0x…40hex) |
| `builder` | Yes | Builder address that will receive fees (lowercase 0x…40hex) |
| `maxFee` | No (default `0.01%`) | Max fee rate the builder may charge per trade |
| `network` | No (default `mainnet`) | `mainnet` or `testnet` |

The `%25` is the URL-encoded `%`. The CLI handles encoding automatically.

## Hosting

The page is one self-contained HTML file. Drop it on any static host.

### Recommended: GitHub Pages from `/docs` (zero config)

If you push this repo to GitHub:

1. Go to your repo's **Settings → Pages**.
2. **Source:** Deploy from a branch.
3. **Branch:** `main`. **Folder:** `/docs`.
4. Save. Your page is live at `https://<user>.github.io/<repo>/setup.html`
   within ~60 seconds.

The empty `docs/.nojekyll` file disables Jekyll processing so the HTML
serves as-is.

### Other static hosts

- **Cloudflare Pages**: `npx wrangler pages deploy ./docs` (free, fast)
- **Vercel**: import the repo, set output directory to `docs/`
- **Netlify**: same, drag-and-drop deploy of `docs/`
- **Plain S3 + CloudFront**: works fine

### Update the CLI to point at your hosted URL

The default URL is set in `internal/cli/agent_setup.go`:

```go
const DefaultSetupURL = "https://thisnewmark.github.io/hyperliquid-pp-cli/setup.html"
```

If you fork or rehost, change it to wherever you're hosting and rebuild
with `make build`.

## Local testing

```bash
# Serve locally
python3 -m http.server 8080 --directory docs

# Open with placeholder params (won't actually submit since the agent isn't real)
open "http://localhost:8080/setup.html?agent=0x0000000000000000000000000000000000000001&builder=0xc8f0cd137e28f717a20f810b46926f92978bbcfa&maxFee=0.01%25"
```

## Security notes

- The page includes an "exact message that gets signed" preview for each
  action. Users can inspect what they're approving before clicking Sign.
- The page checks the URL params for valid 0x…40hex format before showing
  the form. Bad input shows a "must be opened from `agent setup`" error.
- All requests go directly from the browser to Hyperliquid's API
  (`api.hyperliquid.xyz` or testnet equivalent). No middleman.
- The page is single-file with no external runtime dependencies. Easy to
  audit; copy it and host your own if you don't trust ours.

## Customizing copy / styling

The HTML is hand-written, not generated. Edit it freely. The copy is
deliberately written for non-technical users — short sentences, no jargon.
The "What CAN / CANNOT do" panel is the load-bearing safety beat; preserve
it across edits.
