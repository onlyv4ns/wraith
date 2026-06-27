package main

import (
	"fmt"
	"log"
	"time"

	wraith "github.com/onlyv4ns/wraith"
)

func main() {
	browser, err := wraith.New(wraith.Options{})
	if err != nil {
		log.Fatal(err)
	}
	defer browser.Close()

	fmt.Printf("Chrome: %s\nUA: %s\n\n", browser.Version(), browser.UserAgent())

	page, err := browser.NewPage()
	if err != nil {
		log.Fatal(err)
	}
	defer page.Close()

	targets := []string{
		"https://bot.sannysoft.com",
		"https://nowsecure.nl",
		"https://www.browserscan.net",
	}

	for _, url := range targets {
		fmt.Printf("Testing %s ...\n", url)
		if err := page.Navigate(url); err != nil {
			log.Printf("  navigate error: %v", err)
			continue
		}
		page.WaitIdle(5 * time.Second)

		title, _ := page.Title()
		current, _ := page.URL()
		fmt.Printf("  title: %s\n  url:   %s\n", title, current)

		page.Screenshot(fmt.Sprintf("shot-%d.png", len(url)))
	}
}
