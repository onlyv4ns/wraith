package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/go-rod/rod/lib/proto"
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

	target := "https://nowsecure.nl"
	fmt.Printf("Navigating to %s\n", target)

	if err := page.Navigate(target); err != nil {
		log.Fatalf("navigate: %v", err)
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	deadline := time.Now().Add(30 * time.Second)

	for time.Now().Before(deadline) {
		time.Sleep(2 * time.Second)
		html, _ := page.HTML()

		if strings.Contains(html, "Verify you are human") || strings.Contains(html, "cf-chl-widget") {
			fmt.Println("Turnstile detected — clicking checkbox...")
			time.Sleep(time.Duration(400+rng.Intn(400)) * time.Millisecond)

			res, _ := page.Eval(`() => {
				const inp = document.querySelector('[id^=cf-chl-widget]')
				if (!inp) return null
				let el = inp.parentElement
				for (let i = 0; i < 8 && el; i++) {
					const r = el.getBoundingClientRect()
					if (r.height > 30 && r.height < 120 && r.width > 50) {
						return JSON.stringify({x:r.x, y:r.y, w:r.width, h:r.height})
					}
					el = el.parentElement
				}
				return null
			}`)

			s := fmt.Sprintf("%v", res.Value)
			fmt.Printf("  Widget rect: %s\n", s)
			var cx, cy float64
			if s != "null" && strings.HasPrefix(s, "{") {
				var r struct {
					X float64 `json:"x"`
					Y float64 `json:"y"`
					H float64 `json:"h"`
				}
				if json.Unmarshal([]byte(s), &r) == nil && (r.X > 0 || r.Y > 0) {
					cx = r.X + 13
					cy = r.Y + r.H/2
				}
			}
			if cx == 0 && cy == 0 {
				fmt.Println("  WARNING: could not find widget position via DOM")
				continue
			}

			fmt.Printf("  Clicking at (%.0f, %.0f)\n", cx, cy)
			humanClick(page.RodPage().Mouse, rng, cx, cy)

			time.Sleep(5 * time.Second)
			continue
		}

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

func humanClick(mouse interface {
	MoveTo(proto.Point) error
	MoveLinear(proto.Point, int) error
	Down(proto.InputMouseButton, int) error
	Up(proto.InputMouseButton, int) error
}, rng *rand.Rand, tx, ty float64) {
	mouse.MoveTo(proto.NewPoint(float64(200+rng.Intn(600)), float64(500+rng.Intn(250))))
	time.Sleep(time.Duration(150+rng.Intn(200)) * time.Millisecond)
	mouse.MoveLinear(proto.NewPoint(tx+float64(rng.Intn(10)-5), ty-float64(15+rng.Intn(15))), 20+rng.Intn(10))
	time.Sleep(time.Duration(70+rng.Intn(80)) * time.Millisecond)
	mouse.MoveLinear(proto.NewPoint(tx+float64(rng.Intn(4)-2), ty+float64(rng.Intn(4)-2)), 4)
	time.Sleep(time.Duration(60+rng.Intn(80)) * time.Millisecond)
	mouse.Down(proto.InputMouseButtonLeft, 1)
	time.Sleep(time.Duration(50+rng.Intn(50)) * time.Millisecond)
	mouse.Up(proto.InputMouseButtonLeft, 1)
}
