package chrome

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
)

var versionRe = regexp.MustCompile(`(\d+\.\d+\.\d+\.\d+)`)

func Version(binPath string) (string, error) {
	profile, err := os.MkdirTemp("", "wraith-version-")
	if err != nil {
		return "", fmt.Errorf("create temporary Chrome profile: %w", err)
	}
	defer os.RemoveAll(profile)

	out, err := exec.Command(binPath, "--version", "--no-startup-window", "--no-first-run", "--no-default-browser-check", "--user-data-dir="+profile).Output()
	if err != nil {
		return "", fmt.Errorf("chrome --version: %w", err)
	}
	m := versionRe.FindString(strings.TrimSpace(string(out)))
	if m == "" {
		return "", fmt.Errorf("could not parse version from: %s", out)
	}
	return m, nil
}

func UserAgent(version string) string {
	major := strings.SplitN(version, ".", 2)[0]
	switch runtime.GOOS {
	case "darwin":
		return fmt.Sprintf(
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%s.0.0.0 Safari/537.36",
			major,
		)
	case "windows":
		return fmt.Sprintf(
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%s.0.0.0 Safari/537.36",
			major,
		)
	default:
		return fmt.Sprintf(
			"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%s.0.0.0 Safari/537.36",
			major,
		)
	}
}
