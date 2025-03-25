package helpers

import (
	"net/url"
	"os"
)

func EnforceHTTP(url string) string {
	if len(url) < 4 || url[:4] != "http" {
		return "http://" + url
	}
	return url
}

func RemoveDomainError(inputURL string) bool {
	if inputURL == os.Getenv("DOMAIN") {
		return false
	}

	parsedURL, err := url.Parse(inputURL)
	if err != nil {
		return false
	}

	// Get only the hostname
	host := parsedURL.Hostname()

	return host != os.Getenv("DOMAIN")
}
