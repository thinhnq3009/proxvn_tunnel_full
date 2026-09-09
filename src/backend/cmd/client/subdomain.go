package main

import (
	"fmt"
	"strings"
)

func normalizeRequestedSubdomain(value string) (string, error) {
	subdomain := strings.ToLower(strings.TrimSpace(value))
	if subdomain == "" {
		return "", nil
	}
	if len(subdomain) > 63 {
		return "", fmt.Errorf("dài quá 63 ký tự")
	}
	if subdomain[0] == '-' || subdomain[len(subdomain)-1] == '-' {
		return "", fmt.Errorf("không được bắt đầu hoặc kết thúc bằng dấu gạch ngang")
	}
	for _, ch := range subdomain {
		if (ch < 'a' || ch > 'z') && (ch < '0' || ch > '9') && ch != '-' {
			return "", fmt.Errorf("chỉ được chứa chữ cái, chữ số và dấu gạch ngang")
		}
	}
	return subdomain, nil
}
