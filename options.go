package wraith

import "time"

type Options struct {
	Headless     bool
	UserAgent    string
	Proxy        string
	ProfileDir   string
	ChromePath   string
	WindowWidth  int
	WindowHeight int
	Timeout      time.Duration
}

func defaultOptions() Options {
	return Options{
		WindowWidth:  1920,
		WindowHeight: 1080,
		Timeout:      30 * time.Second,
	}
}
