package stealth

import (
	_ "embed"

	"github.com/go-rod/rod"
)

//go:embed stealth.js
var stealthJS string

func InjectOnNewDocument(page *rod.Page) error {
	_, err := page.EvalOnNewDocument(stealthJS)
	return err
}
