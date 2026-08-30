package wraith

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/go-rod/rod/lib/proto"
)

const turnstileMatchThreshold = 0.7


func (p *Page) ClickTurnstile(templatePath ...string) (bool, error) {
	if len(templatePath) > 1 {
		return false, fmt.Errorf("expected at most one Turnstile template path")
	}

	template, err := turnstileTemplate(templatePath)
	if err != nil {
		return false, err
	}
	screenshot, err := p.rod.Screenshot(false, nil)
	if err != nil {
		return false, err
	}
	viewport, _, err := image.Decode(bytes.NewReader(screenshot))
	if err != nil {
		return false, fmt.Errorf("decode viewport screenshot: %w", err)
	}

	point, confidence := templateLocation(viewport, template)
	if confidence < turnstileMatchThreshold {
		domPoint, err := turnstileDOMLocation(p)
		if err != nil {
			return false, err
		}
		if domPoint == nil {
			return false, nil
		}
		return true, p.humanClick(*domPoint, viewport.Bounds())
	}
	return true, p.humanClick(point, viewport.Bounds())
}

func turnstileDOMLocation(p *Page) (*image.Point, error) {
	result, err := p.Eval(`() => {
		const candidates = [
			document.querySelector('[id^=cf-chl-widget]'),
			document.querySelector('iframe[src*="challenges.cloudflare.com"]'),
			document.querySelector('iframe[title*="Cloudflare"]')
		]
		for (const node of candidates) {
			let el = node
			for (let i = 0; i < 8 && el; i++, el = el.parentElement) {
				const r = el.getBoundingClientRect()
				if (r.width > 50 && r.height > 30 && r.height < 150)
					return JSON.stringify({x: r.x, y: r.y, h: r.height})
			}
		}
		return null
	}`)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	raw := fmt.Sprintf("%v", result.Value)
	if raw == "" || raw == "<nil>" || raw == "null" {
		return nil, nil
	}
	var rect struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
		H float64 `json:"h"`
	}
	if err := json.Unmarshal([]byte(raw), &rect); err != nil {
		return nil, nil
	}
	return &image.Point{X: int(rect.X) + 13, Y: int(rect.Y + rect.H/2)}, nil
}

func (p *Page) humanClick(point image.Point, bounds image.Rectangle) error {

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	mouse := p.rod.Mouse
	if err := mouse.MoveTo(proto.NewPoint(float64(rng.Intn(max(bounds.Dx(), 1))), float64(rng.Intn(max(bounds.Dy(), 1))))); err != nil {
		return err
	}
	time.Sleep(time.Duration(150+rng.Intn(200)) * time.Millisecond)
	if err := mouse.MoveLinear(proto.NewPoint(float64(point.X+rng.Intn(11)-5), float64(point.Y-15-rng.Intn(15))), 20+rng.Intn(10)); err != nil {
		return err
	}
	time.Sleep(time.Duration(70+rng.Intn(80)) * time.Millisecond)
	if err := mouse.MoveLinear(proto.NewPoint(float64(point.X+rng.Intn(5)-2), float64(point.Y+rng.Intn(5)-2)), 4); err != nil {
		return err
	}
	time.Sleep(time.Duration(60+rng.Intn(80)) * time.Millisecond)
	if err := mouse.Down(proto.InputMouseButtonLeft, 1); err != nil {
		return err
	}
	time.Sleep(time.Duration(50+rng.Intn(50)) * time.Millisecond)
	if err := mouse.Up(proto.InputMouseButtonLeft, 1); err != nil {
		return err
	}
	return nil
}

func turnstileTemplate(path []string) (image.Image, error) {
	var data []byte
	var err error
	if len(path) == 1 && path[0] != "" {
		data, err = os.ReadFile(path[0])
		if err != nil {
			return nil, fmt.Errorf("read Turnstile template: %w", err)
		}
	} else {
		data, err = base64.StdEncoding.DecodeString(strings.TrimSpace(defaultTurnstileTemplate))
		if err != nil {
			return nil, fmt.Errorf("decode built-in Turnstile template: %w", err)
		}
	}
	template, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode Turnstile template image: %w", err)
	}
	return template, nil
}

func templateLocation(screen, template image.Image) (image.Point, float64) {
	scale := 4
	for screen.Bounds().Dx()/scale > 480 {
		scale++
	}
	s, sw, sh := grayscale(screen, scale)
	t, tw, th := grayscale(template, scale)
	if tw == 0 || th == 0 || tw > sw || th > sh {
		return image.Point{}, -1
	}

	bestPoint, best := matchGrayTemplate(s, sw, sh, t, tw, th, scale)
	inverted := make([]float64, len(t))
	for i, value := range t {
		inverted[i] = 255 - value
	}
	invertedPoint, invertedScore := matchGrayTemplate(s, sw, sh, inverted, tw, th, scale)
	if invertedScore > best {
		return invertedPoint, invertedScore
	}
	return bestPoint, best
}

func matchGrayTemplate(screen []float64, sw, sh int, template []float64, tw, th, scale int) (image.Point, float64) {
	n := float64(tw * th)
	var templateSum, templateSquares float64
	for _, value := range template {
		templateSum += value
		templateSquares += value * value
	}
	templateVariance := templateSquares - templateSum*templateSum/n
	if templateVariance <= 0 {
		return image.Point{}, -1
	}

	best := -1.0
	bestPoint := image.Point{}
	for y := 0; y <= sh-th; y++ {
		for x := 0; x <= sw-tw; x++ {
			var sum, squares, product float64
			for ty := 0; ty < th; ty++ {
				screenRow := (y+ty)*sw + x
				templateRow := ty * tw
				for tx := 0; tx < tw; tx++ {
					sv := screen[screenRow+tx]
					tv := template[templateRow+tx]
					sum += sv
					squares += sv * sv
					product += sv * tv
				}
			}
			variance := squares - sum*sum/n
			if variance <= 0 {
				continue
			}
			score := (product - sum*templateSum/n) / math.Sqrt(variance*templateVariance)
			if score > best {
				best = score
				bestPoint = image.Pt((x+tw/2)*scale, (y+th/2)*scale)
			}
		}
	}
	return bestPoint, best
}

func grayscale(src image.Image, scale int) ([]float64, int, int) {
	bounds := src.Bounds()
	width := bounds.Dx() / scale
	height := bounds.Dy() / scale
	pixels := make([]float64, width*height)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, _ := src.At(bounds.Min.X+x*scale+scale/2, bounds.Min.Y+y*scale+scale/2).RGBA()
			pixels[y*width+x] = 0.299*float64(r>>8) + 0.587*float64(g>>8) + 0.114*float64(b>>8)
		}
	}
	return pixels, width, height
}

const defaultTurnstileTemplate = `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`
