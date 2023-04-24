package httpx

import (
	"errors"
	"net/url"
)

func ValidateProxyURL(proxy string) (string, error) {
	if url1, err := url.Parse(proxy); err == nil && isSupportedProtocol(url1.Scheme) {
		return url1.Scheme, nil
	}
	return "", errors.New("invalid proxy format (It should be http[s]/socks5://[username:password@]host:port)")
}

// isSupportedProtocol checks given protocols are supported
func isSupportedProtocol(value string) bool {
	return value == "http" || value == "https" || value == "socks5"
}
