# Wraith

Go library for undetected browser automation. Bypasses Cloudflare bot protection, Turnstile challenges, and other bot detection systems by combining Chrome-matching fingerprints at every layer — TLS, HTTP/2, browser flags, and JavaScript.

## How it works

Most automation tools are caught because they leave traces at multiple layers:

- **TLS/JA3** — default Go TLS doesn't match Chrome's cipher suite order
- **HTTP/2 SETTINGS** — Go's http2 sends different SETTINGS frames than Chrome
- **Chrome flags** — go-rod and Playwright inject 20+ Puppeteer-style flags that Cloudflare fingerprints (`--disable-background-networking`, `--metrics-recording-only`, etc.)
- **JavaScript** — `navigator.webdriver`, `window.chrome`, `permissions.query`, canvas, WebGL, and audio fingerprints expose automation

Wraith removes all of these signals.

## Features

- Automatic Cloudflare Turnstile bypass with human-like mouse movement
- uTLS transport — JA3/JA4 fingerprint matches real Chrome (`HelloChrome_Auto`)
- HTTP/2 with Chrome-matching SETTINGS frames
- Stealth JS injected before any page script via CDP `addScriptToEvaluateOnNewDocument`
- All go-rod automation flags stripped at launch
- `navigator.webdriver` handled natively by `--disable-blink-features=AutomationControlled` (genuine `[native code]` getter, not overridable by JS)

## Requirements

- Go 1.21+
- Google Chrome or Chromium installed (auto-downloaded if missing)
- Linux: X11 display (`DISPLAY` env set)

## Install

```bash
go get github.com/onlyv4ns/wraith
```

## Usage

### Library

```go
package main

import (
    "fmt"
    "log"

    wraith "github.com/onlyv4ns/wraith"
)

func main() {
    browser, err := wraith.New(wraith.Options{})
    if err != nil {
        log.Fatal(err)
    }
    defer browser.Close()

    page, err := browser.NewPage()
    if err != nil {
        log.Fatal(err)
    }
    defer page.Close()

    page.Navigate("https://example.com")

    title, _ := page.Title()
    fmt.Println(title)

    page.Screenshot("result.png")
}
```

### CLI

```bash
go run ./cmd/wraith -url https://example.com -screenshot out.png
go run ./cmd/wraith -url https://example.com -headless -proxy http://user:pass@host:port
```

### Options

```go
wraith.Options{
    Headless:    false,                    // headless mode (higher detection risk)
    UserAgent:   "",                       // override auto-detected UA
    Proxy:       "http://host:port",       // HTTP/HTTPS proxy
    ProfileDir:  "/path/to/profile",       // persist cookies/session across runs
    ChromePath:  "/path/to/chrome",        // override Chrome binary path
    WindowWidth: 1920,                     // default 1920
    WindowHeight: 1080,                    // default 1080
    Timeout:     30 * time.Second,         // page load timeout
}
```

### HTTP client

```go
client := browser.HTTPClient()
resp, err := client.Get("https://example.com")
```

The HTTP client uses the same uTLS transport as the browser — useful for direct API calls that need matching TLS fingerprints.

## Cloudflare Turnstile

The `examples/cloudflare` example shows automatic Turnstile bypass:

```bash
go run ./examples/cloudflare
```

It polls for the Turnstile widget, locates the checkbox via DOM, and clicks it with randomized human-like mouse movement (approach curve, overshoot, settle, variable hold time).

## Project structure

```
wraith/
├── browser.go              # Browser — launch, stealth flags, NewPage
├── page.go                 # Page — Navigate, Click, Fill, Screenshot, ...
├── options.go              # Options struct
├── cmd/wraith/             # CLI binary
├── examples/
│   ├── basic/              # bot.sannysoft.com, nowsecure.nl, browserscan.net
│   └── cloudflare/         # Cloudflare Turnstile bypass (nowsecure.nl)
└── internal/
    ├── chrome/             # Chrome binary resolution + UA generation
    ├── stealth/            # stealth.js — JS fingerprint patches
    └── transport/          # uTLS + HTTP/2 transport
```
