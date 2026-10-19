# tiny-magpie

A smaller Magpie: a multi-provider model gateway with its original protocol adapters, subscription/provider plugins and desktop/Web GUI.

Tiny keeps upstream model routing and removes Agent management, content libraries, session browsing, sync, intent routing and extra media/search/MCP services. See [the scope](docs/tiny-scope.md) and [upstream maintenance](UPSTREAM.md). It is an experimental fork; validation and remaining acceptance work are recorded in [compatibility](docs/compatibility.md).

Download from [Releases](https://github.com/caelis-labs/tiny-magpie/releases): macOS Apple Silicon DMG, Windows `.exe` and Linux executables (amd64/arm64). Standalone headless CLI binaries cover those same five platform pairs. macOS artifacts are Developer ID signed and notarized; Windows/Linux are unsigned. Intel macOS and Android downloads are outside this fork's release matrix. Verify `SHA256SUMS` before use. See [installation and platform details](docs/releasing.md).

The macOS image shows the app → Applications drag target. Launch the installed copy from Applications. The desktop app checks **Tiny's own releases** every six hours; Settings → General offers preview/stable channels. Download and restart require your confirmation. Version 0.1.2 needs one manual upgrade to gain this updater. See the [graded feature review](docs/feature-review-2026-10-19.md) before the next pruning pass.

## Build and run

Use the Go version in `go.mod`. From this checkout:

```sh
GOWORK=off make cli
./tiny-magpie serve
./tiny-magpie web --no-open
```

Desktop builds use upstream Wails/platform WebViews (`make app` on macOS; `make release-windows` and `make release-linux` on their platforms). The Web GUI is embedded in headless builds too. `script/build_and_run.sh` builds and opens the macOS app with a project-local sandbox profile and gateway port 13425.

Add any supported preset or custom provider using the GUI or CLI; for example:

```sh
./tiny-magpie provider add deepseek <your-key>
./tiny-magpie models
./tiny-magpie plugin add <provider-package>
```

Clients use the original Magpie model ID and API contract, e.g. Base URL `http://127.0.0.1:3425/v1`, Responses/Chat, Messages or Gemini. The loopback token remains `magpie`; remote access uses named gateway keys and existing network guards. Config and caches are independent of the original application.

## Verify

```sh
GOWORK=off go test -tags nogui ./...
python3 scripts/check-scope.py
npm ci
npx playwright install chromium webkit
npm run test:ui
docker build -t tiny-magpie:local-validation .
./scripts/check-container.sh tiny-magpie:local-validation
python3 scripts/verify-gateway.py --baseline /path/to/pinned-magpie --tiny ./tiny-magpie
```

The differential harness runs both binaries with temporary profiles, fake credentials and a loopback fixture upstream. It checks discovery, four-by-four protocol routing with HTTP/SSE, tools/screenshots, both vendor requests and client replies, and removed endpoints. It cannot establish real-vendor or native-window acceptance. Never test against a user's running Magpie or real Agent configuration.

Upstream source remains MIT licensed; retain [LICENSE](LICENSE) and the original contributor history.
