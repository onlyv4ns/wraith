# PRD — Stealth Browser Automation (Go)

**Goal:** Library Go untuk browser automation yang bypass Cloudflare dan bot detection modern, lebih robust dari undetected-chromedriver dengan menyerang semua 4 layer deteksi sekaligus.

---

## Problem Statement

| Library | TLS Fingerprint | HTTP/2 FP | JS Stealth | Binary Patch |
|---------|----------------|-----------|------------|--------------|
| Playwright | ✗ ketahuan | ✗ | Partial | ✗ |
| Selenium UC | ✗ | ✗ | ✓ | ✓ |
| **Target** | ✓ | ✓ | ✓ | ✓ |

Cloudflare Bot Management cek semua 4 layer. Library existing hanya fix 1-2.

---

## Nama Project

`ghostdriver` — Go library + optional CLI.

---

## Scope

**In scope:**
- TLS fingerprint spoofing (JA3/JA4 = real Chrome)
- HTTP/2 fingerprint spoofing (SETTINGS frame, WINDOW_UPDATE order)
- ChromeDriver binary patching (hapus `cdc_` dan detection strings)
- Stealth JS injection sebelum page load
- Browser automation API (navigate, click, fill, screenshot, eval)
- Chrome binary auto-download dan management
- Cookie/session persistence

**Out of scope (v1):**
- CAPTCHA solving (integrasi 2captcha/anticaptcha → v2)
- Proxy rotation built-in (user inject sendiri)
- Firefox support
- Mobile emulation

---

## Arsitektur

```
ghostdriver/
├── cmd/
│   └── ghostdriver/        # CLI entrypoint (opsional)
│       └── main.go
├── internal/
│   ├── patcher/            # ChromeDriver binary patching
│   │   └── patcher.go
│   ├── fingerprint/        # TLS + HTTP/2 fingerprint profiles
│   │   ├── tls.go          # uTLS profiles (Chrome 120, 124, dll)
│   │   └── http2.go        # HTTP/2 SETTINGS frame spoofing
│   ├── stealth/            # JS injection scripts
│   │   └── stealth.go      # embed stealth.js, inject pre-load
│   └── chrome/             # Chrome binary management
│       ├── download.go     # Auto-download Chrome + ChromeDriver
│       └── version.go      # Version matching logic
├── browser.go              # Public API: New(), Navigate(), dll
├── page.go                 # Page methods: Click(), Fill(), Eval(), dll
├── options.go              # BrowserOptions struct
├── stealth/
│   └── stealth.js          # Embedded stealth script (navigator patches)
├── go.mod
├── go.sum
└── examples/
    ├── basic/main.go
    └── cloudflare/main.go
```

---

## Layer Deteksi & Solusi

### Layer 1 — TLS Fingerprint (KRITIS)

**Problem:** Go's `crypto/tls` menghasilkan JA3 hash berbeda dari Chrome real.

**Solusi:** `refraction-networking/utls` dengan preset `HelloChrome_Auto`.

```go
// internal/fingerprint/tls.go
tlsConfig := utls.Config{...}
conn := utls.UClient(rawConn, &tlsConfig, utls.HelloChrome_Auto)
```

Chrome 120 JA3: `771,4865-4866-4867-49195-...,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-17513,29-23-24,0`

### Layer 2 — HTTP/2 Fingerprint

**Problem:** Go's `net/http` mengirim HTTP/2 SETTINGS frame dengan order default, berbeda dari Chrome.

**Solusi:** `fhttp` (fork of net/http) dengan custom SETTINGS:

```go
// Chrome 120 HTTP/2 fingerprint
settings := []http2.Setting{
    {ID: http2.SettingHeaderTableSize, Val: 65536},
    {ID: http2.SettingMaxConcurrentStreams, Val: 1000},
    {ID: http2.SettingInitialWindowSize, Val: 6291456},
    {ID: http2.SettingMaxHeaderListSize, Val: 262144},
}
windowUpdate := 15663105  // Chrome specific
```

### Layer 3 — ChromeDriver Binary Patch

**Problem:** ChromeDriver binary mengandung string `$cdc_` yang terdeteksi via JS.

**Solusi:** Patch binary sebelum launch:

```go
// internal/patcher/patcher.go
func Patch(driverPath string) error {
    data, _ := os.ReadFile(driverPath)
    // Replace cdc_ variable name dengan random string panjang sama
    patched := replaceCDCVar(data)
    // Replace devtools_active_port string
    patched = replaceDevtoolsString(patched)
    return os.WriteFile(driverPath, patched, 0755)
}
```

### Layer 4 — JavaScript Stealth

**Problem:** `navigator.webdriver = true`, `window.chrome` missing, permissions behavior aneh.

**Solusi:** Inject stealth.js via CDP `Page.addScriptToEvaluateOnNewDocument` sebelum page load:

```js
// stealth/stealth.js (di-embed ke binary)
Object.defineProperty(navigator, 'webdriver', {get: () => undefined})
window.chrome = { runtime: {} }
// Fix navigator.plugins (minimal 3 plugins)
// Fix navigator.languages
// Fix Notification.permission behavior
// Fix iframe contentWindow.navigator.webdriver
```

---

## Public API

```go
// Minimal, mirip Playwright tapi lebih simpel
browser, err := ghostdriver.New(ghostdriver.Options{
    Headless:    false,       // true = lebih mudah detect
    UserAgent:   "",          // auto dari Chrome version
    Proxy:       "",          // "http://user:pass@host:port"
    ProfileDir:  "",          // persist cookies/session
    ChromePath:  "",          // auto-detect atau custom path
})
defer browser.Close()

page, err := browser.NewPage()
err = page.Navigate("https://nowsecure.nl")

// Tunggu elemen
err = page.WaitForSelector("#result", 10*time.Second)

// Interact
err = page.Click("#button")
err = page.Fill("#input", "text")

// Extract
text, err := page.Text("#element")
html, err := page.HTML()
cookies, err := page.Cookies()

// Raw eval
result, err := page.Eval(`document.title`)

// Screenshot
err = page.Screenshot("shot.png")
```

---

## Chrome Version Management

```
chrome/
  └── download.go
```

- Auto-detect Chrome yang sudah install di sistem
- Jika tidak ada: download Chrome for Testing dari `googlechromelabs.github.io/chrome-for-testing/`
- Match ChromeDriver version dengan Chrome version (major version harus sama)
- Cache di `~/.ghostdriver/chrome/`

---

## Dependencies

```go
// go.mod
require (
    github.com/refraction-networking/utls   v1.6.x  // TLS fingerprint
    github.com/Danny-Dasilva/fhttp         v1.x    // HTTP/2 fingerprint  
    github.com/go-rod/rod                  v0.x    // CDP browser control
)
```

> Tidak ada dependency lain. Total ~3 deps.

---

## Detection Test Targets

Sebelum release, harus pass semua:

| URL | Test |
|-----|------|
| `https://nowsecure.nl` | TLS + JS fingerprint |
| `https://bot.sannysoft.com` | Comprehensive bot check |
| `https://fingerprint.com/demo` | Browser fingerprint |
| `https://www.browserscan.net` | TLS JA3/JA4 |
| Cloudflare-protected site | Real-world test |

---

## Milestones

| Phase | Scope | Estimasi |
|-------|-------|----------|
| **P1** | Patcher + basic Rod wrapper + stealth JS | 3-4 hari |
| **P2** | uTLS integration + HTTP/2 FP | 3-4 hari |
| **P3** | Chrome auto-download + version management | 2 hari |
| **P4** | Public API polish + examples | 2 hari |
| **P5** | Pass semua detection tests | ongoing |

---

## Non-Goals

- Bukan pengganti Playwright untuk test UI biasa
- Bukan tool untuk DDoS atau fraud
- Tidak menyembunyikan IP (gunakan proxy/VPN sendiri)

---

## Referensi

- [undetected-chromedriver](https://github.com/ultrafunkamsterdam/undetected-chromedriver) — inspirasi, Python
- [refraction-networking/utls](https://github.com/refraction-networking/utls) — TLS spoofing
- [Danny-Dasilva/CycleTLS](https://github.com/Danny-Dasilva/CycleTLS) — TLS + HTTP/2 in Go
- [go-rod/rod](https://github.com/go-rod/rod) — CDP automation
- [Kaliiiiiiiiii/brotector](https://github.com/Kaliiiiiiiiii/brotector) — detection test tool
