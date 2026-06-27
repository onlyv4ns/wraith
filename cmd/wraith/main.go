package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	wraith "github.com/yourusername/wraith"
)

func main() {
	url := flag.String("url", "", "URL to visit")
	headless := flag.Bool("headless", false, "Run headless")
	screenshot := flag.String("screenshot", "", "Save screenshot to path")
	proxy := flag.String("proxy", "", "Proxy URL (http://user:pass@host:port)")
	profile := flag.String("profile", "", "Profile directory for session persistence")
	timeout := flag.Duration("timeout", 30*time.Second, "Page load timeout")
	flag.Parse()

	if *url == "" {
		fmt.Fprintln(os.Stderr, "usage: wraith -url <url> [options]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	browser, err := wraith.New(wraith.Options{
		Headless:   *headless,
		Proxy:      *proxy,
		ProfileDir: *profile,
		Timeout:    *timeout,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer browser.Close()

	fmt.Printf("Chrome %s\n", browser.Version())

	page, err := browser.NewPage()
	if err != nil {
		log.Fatal(err)
	}
	defer page.Close()

	if err := page.Navigate(*url); err != nil {
		log.Fatalf("navigate: %v", err)
	}
	page.WaitIdle(5 * time.Second)

	title, _ := page.Title()
	current, _ := page.URL()
	html, _ := page.HTML()
	fmt.Printf("Title: %s\nURL:   %s\nSize:  %d bytes\n", title, current, len(html))

	if *screenshot != "" {
		if err := page.Screenshot(*screenshot); err != nil {
			log.Printf("screenshot error: %v", err)
		} else {
			fmt.Printf("Screenshot: %s\n", *screenshot)
		}
	}
}
