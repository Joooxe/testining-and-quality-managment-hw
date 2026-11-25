package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"
)

const RequestTimeout = 5 * time.Second

func BaseURL() (string, error) {
	raw := strings.TrimSpace(os.Getenv("API_BASE_URL"))
	if raw == "" {
		raw = "http://localhost:8080"
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("parse API_BASE_URL: %w", err)
	}
	if (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Hostname() == "" ||
		parsed.User != nil || parsed.ForceQuery || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", fmt.Errorf("API_BASE_URL must be an HTTP(S) URL without credentials, query or fragment")
	}
	return strings.TrimRight(parsed.String(), "/"), nil
}
