package wraith

import (
	"image"
	"image/color"
	"image/draw"
	"testing"
)

func TestTemplateLocation(t *testing.T) {
	template := image.NewGray(image.Rect(0, 0, 32, 24))
	for y := 0; y < 24; y++ {
		for x := 0; x < 32; x++ {
			template.SetGray(x, y, color.Gray{Y: uint8((x*37 + y*71 + x*y) % 256)})
		}
	}
	screen := image.NewGray(image.Rect(0, 0, 96, 72))
	draw.Draw(screen, screen.Bounds(), &image.Uniform{C: color.Gray{Y: 10}}, image.Point{}, draw.Src)
	draw.Draw(screen, image.Rect(40, 28, 72, 52), template, image.Point{}, draw.Src)

	point, confidence := templateLocation(screen, template)
	if point != (image.Pt(56, 40)) || confidence < 0.99 {
		t.Fatalf("got point %v with confidence %.3f", point, confidence)
	}

	darkScreen := image.NewGray(image.Rect(0, 0, 96, 72))
	draw.Draw(darkScreen, darkScreen.Bounds(), &image.Uniform{C: color.Gray{Y: 245}}, image.Point{}, draw.Src)
	darkTemplate := image.NewGray(template.Bounds())
	for y := 0; y < 24; y++ {
		for x := 0; x < 32; x++ {
			darkTemplate.SetGray(x, y, color.Gray{Y: 255 - template.GrayAt(x, y).Y})
		}
	}
	draw.Draw(darkScreen, image.Rect(40, 28, 72, 52), darkTemplate, image.Point{}, draw.Src)
	darkPoint, darkConfidence := templateLocation(darkScreen, template)
	if darkPoint != (image.Pt(56, 40)) || darkConfidence < 0.99 {
		t.Fatalf("dark theme: got point %v with confidence %.3f", darkPoint, darkConfidence)
	}
}

func TestBuiltInTurnstileTemplate(t *testing.T) {
	if _, err := turnstileTemplate(nil); err != nil {
		t.Fatal(err)
	}
}
