package domain

import "strings"

func Match(host string, domains []string) bool {
	if len(domains) == 0 {
		return false
	}

	for _, v := range domains {
		if host == v || strings.HasSuffix(host, "."+v) {
			return true
		}
	}

	return false
}
