package chrome

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/go-rod/rod/lib/launcher"
)

func ResolveChrome(customPath string) (string, error) {
	if customPath != "" {
		if _, err := os.Stat(customPath); err != nil {
			return "", fmt.Errorf("chrome not found at: %s", customPath)
		}
		return customPath, nil
	}

	if path := systemChromePath(); path != "" {
		return path, nil
	}

	path, err := launcher.NewBrowser().Get()
	if err != nil {
		return "", fmt.Errorf("auto-download chrome: %w", err)
	}
	return path, nil
}

func systemChromePath() string {
	for _, c := range systemCandidates() {
		if p, err := exec.LookPath(c); err == nil {
			return p
		}
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return ""
}

func systemCandidates() []string {
	switch runtime.GOOS {
	case "linux":
		return []string{"google-chrome", "google-chrome-stable", "chromium", "chromium-browser"}
	case "darwin":
		return []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
		}
	case "windows":
		return []string{
			`C:\Program Files\Google\Chrome\Application\chrome.exe`,
			`C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`,
		}
	}
	return nil
}
