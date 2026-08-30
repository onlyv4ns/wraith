package wraith

import (
	"fmt"
	"net/http"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/onlyv4ns/wraith/internal/chrome"
	"github.com/onlyv4ns/wraith/internal/stealth"
	"github.com/onlyv4ns/wraith/internal/transport"
)

type Browser struct {
	rod     *rod.Browser
	opts    Options
	version string
	ua      string
}

func New(opts Options) (*Browser, error) {
	o := defaultOptions()
	if opts.Headless {
		o.Headless = true
	}
	if opts.UserAgent != "" {
		o.UserAgent = opts.UserAgent
	}
	if opts.Proxy != "" {
		o.Proxy = opts.Proxy
	}
	if opts.ProfileDir != "" {
		o.ProfileDir = opts.ProfileDir
	}
	if opts.ChromePath != "" {
		o.ChromePath = opts.ChromePath
	}
	if opts.WindowWidth > 0 {
		o.WindowWidth = opts.WindowWidth
	}
	if opts.WindowHeight > 0 {
		o.WindowHeight = opts.WindowHeight
	}
	if opts.Timeout > 0 {
		o.Timeout = opts.Timeout
	}

	chromePath, err := chrome.ResolveChrome(o.ChromePath)
	if err != nil {
		return nil, fmt.Errorf("resolve chrome: %w", err)
	}

	version, _ := chrome.Version(chromePath)

	ua := o.UserAgent
	if ua == "" {
		if version != "" {
			ua = chrome.UserAgent(version)
		} else {
			ua = chrome.UserAgent("120")
		}
	}

	l := launcher.New().
		Bin(chromePath).
		Leakless(false).
		Headless(o.Headless).
		Delete("disable-background-networking").
		Delete("disable-background-timer-throttling").
		Delete("disable-backgrounding-occluded-windows").
		Delete("disable-breakpad").
		Delete("disable-client-side-phishing-detection").
		Delete("disable-component-extensions-with-background-pages").
		Delete("disable-default-apps").
		Delete("disable-hang-monitor").
		Delete("disable-ipc-flooding-protection").
		Delete("disable-popup-blocking").
		Delete("disable-prompt-on-repost").
		Delete("disable-renderer-backgrounding").
		Delete("disable-sync").
		Delete("disable-site-isolation-trials").
		Delete("enable-automation").
		Delete("enable-features").
		Delete("disable-features").
		Delete("force-color-profile").
		Delete("metrics-recording-only").
		Delete("use-mock-keychain").
		Set("disable-blink-features", "AutomationControlled").
		Set("exclude-switches", "enable-automation").
		Set("start-maximized", "").
		Set("user-agent", ua).
		Set("no-sandbox", "").
		Set("ozone-platform", "x11")

	if o.Proxy != "" {
		l = l.Proxy(o.Proxy)
	}
	if o.ProfileDir != "" {
		l = l.UserDataDir(o.ProfileDir)
	}

	url, err := l.Launch()
	if err != nil {
		return nil, fmt.Errorf("launch chrome: %w", err)
	}

	browser := rod.New().ControlURL(url)
	if err := browser.Connect(); err != nil {
		return nil, fmt.Errorf("connect to chrome: %w", err)
	}

	return &Browser{rod: browser, opts: o, version: version, ua: ua}, nil
}

func (b *Browser) NewPage() (*Page, error) {
	p, err := b.rod.Page(proto.TargetCreateTarget{URL: "about:blank"})
	if err != nil {
		return nil, err
	}

	if err := stealth.InjectOnNewDocument(p); err != nil {
		return nil, fmt.Errorf("stealth inject: %w", err)
	}
	if err := p.SetUserAgent(&proto.NetworkSetUserAgentOverride{UserAgent: b.ua}); err != nil {
		return nil, fmt.Errorf("set user agent: %w", err)
	}

	return &Page{rod: p}, nil
}

func (b *Browser) HTTPClient() *http.Client {
	return transport.NewClient()
}

func (b *Browser) RodBrowser() *rod.Browser { return b.rod }
func (b *Browser) Version() string          { return b.version }
func (b *Browser) UserAgent() string        { return b.ua }
func (b *Browser) Close() error             { return b.rod.Close() }
