package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	wraith "github.com/onlyv4ns/wraith"
)

func main() {
	headless := flag.Bool("headless", false, "Run headless")
	flag.Parse()

	browser, err := wraith.New(wraith.Options{Headless: *headless})
	if err != nil {
		log.Fatal(err)
	}
	defer browser.Close()

	page, err := browser.NewPage()
	if err != nil {
		log.Fatal(err)
	}
	defer page.Close()

	target := "https://nowsecure.nl/"
	fmt.Printf("Navigating to %s\n", target)

	if err := page.Navigate(target); err != nil {
		log.Fatalf("navigate: %v", err)
	}

	deadline := time.Now().Add(30 * time.Second)

	for time.Now().Before(deadline) {
		time.Sleep(2 * time.Second)
		clicked, err := page.ClickTurnstile()
		if err != nil {
			log.Printf("Turnstile detection: %v", err)
		} else if clicked {
			fmt.Println("Turnstile detected — checkbox clicked")
			time.Sleep(5 * time.Second)
			continue
		}

		html, _ := page.HTML()
		passed := !strings.Contains(html, "Just a moment") &&
			!strings.Contains(html, "security verification") &&
			!strings.Contains(html, "Verifying you are human") &&
			!strings.Contains(html, "Performing security verification")

		if passed {
			title, _ := page.Title()
			url, _ := page.URL()
			fmt.Printf("\n✓ Cloudflare bypassed!\nTitle: %s\nURL:   %s\n", title, url)
			page.Screenshot("cloudflare-result.png")
			fmt.Println("Screenshot saved: cloudflare-result.png")
			return
		}

		fmt.Printf("  Waiting... (%ds left)\n", int(time.Until(deadline).Seconds()))
	}

	fmt.Println("✗ Timed out")
	page.Screenshot("cloudflare-result.png")
}
