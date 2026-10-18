# tiny-magpie

A smaller Magpie: a multi-provider model gateway with its original protocol adapters, subscription/provider plugins and desktop/Web GUI.

Tiny keeps upstream model routing and removes Agent management, content libraries, session browsing, sync, intent routing and extra media/search/MCP services. See [the scope](docs/tiny-scope.md) and [upstream maintenance](UPSTREAM.md). It is an experimental fork; validation and remaining acceptance work are recorded in [compatibility](docs/compatibility.md).

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
python3 scripts/verify-gateway.py --baseline /path/to/pinned-magpie --tiny ./tiny-magpie
```

The differential harness runs both binaries with temporary profiles, fake credentials and a loopback fixture upstream. It checks discovery, four-by-four protocol routing with HTTP/SSE and removed endpoints. It cannot establish real-vendor or native-window acceptance. Never test against a user's running Magpie or real Agent configuration.

Upstream source remains MIT licensed; retain [LICENSE](LICENSE) and the original contributor history.
